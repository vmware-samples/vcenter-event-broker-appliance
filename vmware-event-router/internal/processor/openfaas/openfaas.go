package openfaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/embano1/waitgroup"
	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	defaultTopicDelimiter  = ","
	defaultRebuildInterval = time.Second * 10
	defaultTimeout         = time.Second * 15
	waitShutdown           = 5 * time.Second // wait for processing to finish during shutdown
)

var (
	// ErrStopped error is returned when the processor has been shutdown but there
	// are still inflight processing requests
	ErrStopped = errors.New("processor already stopped")

	// assert we implement Processor interface
	_ processor.Processor = (*Processor)(nil)
)

// invokeFunc is a function which invokes the given OpenFaaS function with the
// specified message body. It returns the function response message, status
// code, http headers and error.
type invokeFunc func(ctx context.Context, function string, message []byte) ([]byte, int, http.Header, error)

// waitFunc is a wait function passed to waitForAll
type waitFunc func(ctx context.Context) error

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
	respChan chan ofsdk.InvokerResponse // responses from sync fn invocation
	wg       waitgroup.WaitGroup        // used in graceful shutdown

	// options
	topicDelimiter  string
	rebuildInterval time.Duration
	gatewayTimeout  time.Duration
	logger.Logger

	lock    sync.RWMutex
	stats   metrics.EventStats
	stopped bool
}

// NewProcessor returns an OpenFaaS processor for the given configuration.
func NewProcessor(ctx context.Context, cfg *config.ProcessorConfigOpenFaaS, ms metrics.Receiver, log logger.Logger, opts ...Option) (*Processor, error) {
	ofLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		proc := strings.ToUpper(string(config.ProcessorOpenFaaS))
		ofLog = zapSugared.Named(fmt.Sprintf("[%s]", proc))
	}

	ofProcessor := Processor{
		topicDelimiter:  defaultTopicDelimiter,
		rebuildInterval: defaultRebuildInterval,
		gatewayTimeout:  defaultTimeout,
		Logger:          ofLog,
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
		ofLog.Warn("disabling basic auth: no authentication data provided")
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
		PrintSync:                true,
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
		// we don't get a callback for async invocations
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

// Process implements the stream processor interface and invokes any OpenFaaS
// function subscribed to the passed cloud event. If the processor has already
// been shutdown, ErrStopped will be returned.
func (p *Processor) Process(ctx context.Context, ce cloudevents.Event) error {
	if p.isStopped() {
		return ErrStopped
	}

	// coordinate concurrent shutdown
	p.wg.Add(1)
	defer p.wg.Done()

	p.Debugw("processing event", "eventID", ce.ID(), "event", ce)
	topic, message, err := handleEvent(ce)
	if err != nil {
		return processor.NewError(config.ProcessorOpenFaaS, errors.Wrapf(err, "handle event %v", ce))
	}

	p.Debugw("created new outbound event for subscribers", "message", string(message))
	p.Infow("invoking function(s) for event", "eventID", ce.ID(), "topic", topic)
	defer func() {
		p.Infow("finished processing of event", "eventID", ce.ID(), "topic", topic)
	}()

	m, err := p.controller.InvokeWithContext(ctx, topic, message)
	if err != nil {
		return processor.NewError(config.ProcessorOpenFaaS, errors.Wrap(err, "invoke function"))
	}

	p.Infow("function(s) matched for event", "count", m, "eventID", ce.ID(), "topic", topic)
	if m == 0 {
		return nil
	}

	p.Infow("waiting for functions to return", "count", m)

	waitFn := waitForOne(p.respChan, p.controller.InvokeFunction, message, p.Logger, defaultRetryOpts...)
	return waitForAll(ctx, m, waitFn)
}

// waitForAll waits for waitN wait functions to return. It returns the first
// error encountered by fn or nil on success.
func waitForAll(ctx context.Context, waitN int, fn waitFunc) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// expect m callbacks
	for i := 0; i < waitN; i++ {
		eg.Go(func() error {
			return fn(egCtx)
		})
	}

	// wait for all groups to finish and return with nil or error
	return eg.Wait()
}

// waitForOne waits for one InvokerResponse from resCh from a single function
// invocation and handles retries in case of failure. If the processor has
// already been stopped, ErrStopped will be returned.
func waitForOne(resCh <-chan ofsdk.InvokerResponse, invoker invokeFunc, retryMsg []byte, log logger.Logger, retryOpts ...retry.Option) waitFunc {
	return func(ctx context.Context) error {
		var retryCount int32

		if res, ok := <-resCh; ok {
			// return early
			if isSuccessful(res.Status, res.Error) {
				log.Infow("successfully invoked function", "function", res.Function, "topic", res.Topic, "retries", retryCount)
				return nil
			}

			// retries unless error is nil or of type retry.Unrecoverable
			err := retry.Do(retryFunc(ctx, res, invoker, retryMsg, &retryCount), retryOpts...)
			if err != nil {
				log.Errorw("could not invoke function", "function", res.Function, "topic", res.Topic, "retries", retryCount, "error", err)
				return nil
			}

			log.Infow("successfully invoked function", "function", res.Function, "topic", res.Topic, "retries", retryCount)
			return nil
		}

		// avoid deadlock when processor is stopped concurrently
		return ErrStopped
	}
}

// handleEvent returns the OpenFaaS subscription topic, e.g. VmPoweredOnEvent,
// and outbound JSON-encoded event message in []byte for the given CloudEvent
func handleEvent(event cloudevents.Event) (string, []byte, error) {
	message, err := json.Marshal(event)
	if err != nil {
		return "", nil, errors.Wrapf(err, "JSON-encode CloudEvent %v", event)
	}
	return event.Subject(), message, nil
}

// PushMetrics pushes metrics to the specified metrics receiver
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

// Shutdown performs a clean shutdown of the OpenFaaS processor. It must not be
// called more than once and only after all inflight event processing requests
// have finished to avoid a panic. If the processor has already been stopped
// ErrStopped is returned.
func (p *Processor) Shutdown(_ context.Context) error {
	p.Logger.Infof("attempting graceful shutdown")
	if p.isStopped() {
		return ErrStopped
	}

	p.lock.Lock()
	p.stopped = true
	p.lock.Unlock()

	p.Logger.Infof("waiting up to %v for inflight events to finish processing", waitShutdown)
	err := p.wg.WaitTimeout(waitShutdown)
	if err != nil {
		p.Logger.Warnw("timeout waiting for inflight events", "error", err)
		p.Logger.Warn("enforcing shutdown")
	}

	// free resources - if waitShutdown fails (they're still inflight processor
	// invocations) this might (intentionally) lead to a panic by writing to a
	// closed channel from the worker routines
	close(p.respChan)
	return nil
}

func (p *Processor) isStopped() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.stopped
}
