package openfaas

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	defaultTopicDelimiter  = ","
	defaultRebuildInterval = time.Second * 10
	defaultTimeout         = time.Second * 15
)

type logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// invokeFn is a function which invokes the given OpenFaaS function (fn) with
// the given message body
type invokeFn func(ctx context.Context, fn string, message []byte) ([]byte, int, http.Header, error)

// responseFunc implements ResponseSubscriber and is used to configure the
// default response handler for the OpenFaaS processor
type responseFunc func(ofsdk.InvokerResponse)

// Response is a wrapper function to implement the OpenFaaS Response handler
// interface
func (r responseFunc) Response(res ofsdk.InvokerResponse) {
	r(res)
}

// Processor implements the Processor interface
type Processor struct {
	controller ofsdk.Controller
	ofsdk.ResponseSubscriber
	respChan chan ofsdk.InvokerResponse // retrieve errors from sync fn invocation

	// options
	verbose         bool
	topicDelimiter  string
	rebuildInterval time.Duration
	gatewayTimeout  time.Duration
	// TODO (@embano1): make log interface for all processors/streams
	*log.Logger

	lock  sync.RWMutex
	stats metrics.EventStats
}

// NewProcessor returns an OpenFaaS processor for the given stream
// source. Asynchronous function invocation can be configured for
// high-throughput (non-blocking) requirements.
func NewProcessor(ctx context.Context, cfg *config.ProcessorConfigOpenFaaS, ms metrics.Receiver, opts ...Option) (*Processor, error) {
	// defaults
	l := log.New(os.Stdout, color.Purple("[OpenFaaS] "), log.LstdFlags)
	ofProcessor := Processor{
		topicDelimiter:  defaultTopicDelimiter,
		rebuildInterval: defaultRebuildInterval,
		gatewayTimeout:  defaultTimeout,
		Logger:          l,
		respChan:        make(chan ofsdk.InvokerResponse),
	}
	ofProcessor.ResponseSubscriber = defaultResponseHandler(&ofProcessor)

	// apply options
	for _, opt := range opts {
		opt(&ofProcessor)
	}

	if cfg == nil {
		return nil, errors.New("no OpenFaaS configuration found")
	}

	// it's ok to pass empty credentials to OpenFaaS if basic_auth is not used
	var credentials auth.BasicAuthCredentials

	switch cfg.Auth {
	case nil:
		l.Println("no authentication data provided, disabling basic auth")
	default:
		if cfg.Auth.Type != config.BasicAuth {
			return nil, fmt.Errorf("unsupported authentication method %q specified for this processor", cfg.Auth.Type)
		}

		if cfg.Auth.BasicAuth == nil {
			return nil, errors.New("basic auth credentials must be specified")
		}

		credentials.User = cfg.Auth.BasicAuth.Username
		credentials.Password = cfg.Auth.BasicAuth.Password
	}

	ctlCfg := ofsdk.ControllerConfig{
		GatewayURL:               cfg.Address,
		TopicAnnotationDelimiter: ofProcessor.topicDelimiter,
		RebuildInterval:          ofProcessor.rebuildInterval,
		UpstreamTimeout:          ofProcessor.gatewayTimeout,
		AsyncFunctionInvocation:  cfg.Async,
		PrintSync:                ofProcessor.verbose,
	}

	ctl := ofsdk.NewController(&credentials, &ctlCfg, ofProcessor.Logger)
	ofProcessor.controller = ctl
	ofProcessor.controller.Subscribe(&ofProcessor)
	ofProcessor.controller.BeginMapBuilder()

	// pre-populate the metrics stats
	ofProcessor.stats = metrics.EventStats{
		Provider:    string(config.ProcessorOpenFaaS),
		Type:        config.EventProcessor,
		Address:     cfg.Address,
		Started:     time.Now().UTC(),
		Invocations: make(map[string]*metrics.InvocationDetails),
	}
	go ofProcessor.PushMetrics(ctx, ms)

	return &ofProcessor, nil
}

// defaultResponseHandler records metrics and handles invoker responses
func defaultResponseHandler(of *Processor) responseFunc {
	return func(res ofsdk.InvokerResponse) {
		// TODO: currently we only support metrics when in sync invocation mode because
		// we don't have a callback for async invocations
		of.lock.Lock()

		// check for existing topic entry
		if _, ok := of.stats.Invocations[res.Topic]; !ok {
			of.stats.Invocations[res.Topic] = &metrics.InvocationDetails{}
		}

		// record metrics
		// note: only first invocation result is captured (no retries)
		if isSuccessful(res.Status, res.Error) {
			of.stats.Invocations[res.Topic].Success()
		} else {
			of.stats.Invocations[res.Topic].Failure()
		}
		of.lock.Unlock()

		of.respChan <- res
	}
}

// Process implements the stream processor interface
func (p *Processor) Process(ctx context.Context, ce cloudevents.Event) error {
	if p.verbose {
		p.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	topic, message, err := handleEvent(ce)
	if err != nil {
		return processor.NewError(config.ProcessorOpenFaaS, errors.Wrapf(err, "handle event %v", ce))
	}

	if p.verbose {
		p.Printf("created new outbound event for subscribers: %q", string(message))
	}

	p.Printf("invoking function(s) for event %q on topic: %q", ce.ID(), topic)
	defer func() {
		p.Printf("finished processing of event %q on topic: %q", ce.ID(), topic)
	}()

	m, err := p.controller.InvokeWithContext(ctx, topic, message)
	if err != nil {
		return processor.NewError(config.ProcessorOpenFaaS, errors.Wrap(err, "invoke function"))
	}

	p.Printf("%d function(s) matched for event %q on topic: %q", m, ce.ID(), topic)
	if m == 0 {
		return nil
	}

	p.Printf("waiting for %d functions to return", m)

	return waitForAll(ctx, m, p.respChan, p.controller.InvokeFunction, message, p.Logger)
}

// waitForAll waits for all functions specified in numFn before returning
func waitForAll(ctx context.Context, numFn int, respChan <-chan ofsdk.InvokerResponse, invoker invokeFn, retryMsg []byte, log logger) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// expect m callbacks
	for i := 0; i < numFn; i++ {
		res := <-respChan
		eg.Go(waitFor(egCtx, res, invoker, retryMsg, log))
	}

	// wait for all groups to finish and return error if any, otherwise nil
	return eg.Wait()
}

// waitFor waits for a single function invocation to return and handles retries
// in case of failures
func waitFor(ctx context.Context, res ofsdk.InvokerResponse, invoker invokeFn, retryMsg []byte, log logger) func() error {
	// configure retry options
	retryOps := []retry.Option{
		retry.Attempts(3),
		retry.MaxDelay(5 * time.Second),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		// retry.LastErrorOnly(true),
	}

	return func() error {
		var retryCount int32

		// return early
		if isSuccessful(res.Status, res.Error) {
			log.Printf("successfully invoked function %q for topic %q (retries: %d)", res.Function, res.Topic, retryCount)
			return nil
		}

		// retries unless error is nil or of type retry.Unrecoverable
		err := retry.Do(retryFunc(ctx, res, invoker, retryMsg, &retryCount), retryOps...)
		if err != nil {
			log.Printf("could not invoke function %q for topic %q (retries: %d): %v", res.Function, res.Topic, retryCount, err)
			return nil
		}

		log.Printf("successfully invoked function %q for topic %q (retries: %d)", res.Function, res.Topic, retryCount)
		return nil
	}
}

// handleEvent returns the OpenFaaS subscription topic, e.g. VmPoweredOnEvent,
// and outbound event message ([]byte(CloudEvent) for the given CloudEvent
func handleEvent(event cloudevents.Event) (string, []byte, error) {
	message, err := json.Marshal(event)
	if err != nil {
		return "", nil, errors.Wrapf(err, "JSON-encode CloudEvent %v", event)
	}
	return event.Subject(), message, nil
}

func (p *Processor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.lock.RLock()
			ms.Receive(&p.stats)
			p.lock.RUnlock()
		}
	}
}
