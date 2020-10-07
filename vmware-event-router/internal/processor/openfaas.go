package processor

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	defaultTopicDelimiter  = ","
	defaultRebuildInterval = time.Second * 10
	defaultTimeout         = time.Second * 15
)

// responseFunc implements ResponseSubscriber and is used to configure the
// default response handler for the OpenFaaS processor
type responseFunc func(ofsdk.InvokerResponse)

func (r responseFunc) Response(res ofsdk.InvokerResponse) {
	r(res)
}

// OpenfaasProcessor implements the Processor interface
type OpenfaasProcessor struct {
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

// NewOpenFaaSProcessor returns an OpenFaaS processor for the given stream
// source. Asynchronous function invocation can be configured for
// high-throughput (non-blocking) requirements.
func NewOpenFaaSProcessor(ctx context.Context, cfg *config.ProcessorConfigOpenFaaS, ms metrics.Receiver, opts ...OpenFaaSOption) (*OpenfaasProcessor, error) {
	// defaults
	logger := log.New(os.Stdout, color.Purple("[OpenFaaS] "), log.LstdFlags)
	ofProcessor := OpenfaasProcessor{
		topicDelimiter:  defaultTopicDelimiter,
		rebuildInterval: defaultRebuildInterval,
		gatewayTimeout:  defaultTimeout,
		Logger:          logger,
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

	// it's ok to pass empty creds to OpenFaaS if basic_auth is not used
	var creds auth.BasicAuthCredentials

	switch cfg.Auth {
	case nil:
		logger.Println("no authentication data provided, disabling basic auth")
	default:
		if cfg.Auth.Type != config.BasicAuth {
			return nil, fmt.Errorf("unsupported authentication method %q specified for this processor", cfg.Auth.Type)
		}

		if cfg.Auth.BasicAuth == nil {
			return nil, errors.New("basic auth credentials must be specified")
		}

		creds.User = cfg.Auth.BasicAuth.Username
		creds.Password = cfg.Auth.BasicAuth.Password
	}

	ofconfig := ofsdk.ControllerConfig{
		GatewayURL:               cfg.Address,
		TopicAnnotationDelimiter: ofProcessor.topicDelimiter,
		RebuildInterval:          ofProcessor.rebuildInterval,
		UpstreamTimeout:          ofProcessor.gatewayTimeout,
		AsyncFunctionInvocation:  cfg.Async,
		PrintSync:                ofProcessor.verbose,
	}

	ofcontroller := ofsdk.NewController(&creds, &ofconfig)
	ofProcessor.controller = ofcontroller
	ofProcessor.controller.Subscribe(&ofProcessor)
	ofProcessor.controller.BeginMapBuilder()

	// pre-populate the metrics stats
	ofProcessor.stats = metrics.EventStats{
		Provider:    string(config.ProcessorOpenFaaS),
		Type:        config.EventProcessor,
		Address:     cfg.Address,
		Started:     time.Now().UTC(),
		Invocations: make(map[string]int),
	}
	go ofProcessor.PushMetrics(ctx, ms)

	return &ofProcessor, nil
}

// defaultResponseHandler captures errors caused by invoking or returned by
// functions and prints status information for each function invocation
func defaultResponseHandler(of *OpenfaasProcessor) responseFunc {
	return func(res ofsdk.InvokerResponse) {
		// TODO: currently we only support metrics when in sync invocation mode
		// because we don't have a callback for async invocations
		of.lock.Lock()
		of.stats.Invocations[res.Topic]++
		of.lock.Unlock()

		of.respChan <- res
	}
}

// Process implements the stream processor interface
func (of *OpenfaasProcessor) Process(ce cloudevents.Event) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if of.verbose {
		of.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	topic, message, err := handleEvent(ce)
	if err != nil {
		return processorError(config.ProcessorOpenFaaS, errors.Wrapf(err, "handle event %v", ce))
	}

	if of.verbose {
		of.Printf("created new outbound event for subscribers: %q", string(message))
	}

	of.Printf("invoking function(s) for event %q on topic: %q", ce.ID(), topic)

	m, err := of.controller.InvokeWithContext(ctx, topic, &message)
	if err != nil {
		return processorError(config.ProcessorOpenFaaS, errors.Wrap(err, "invoke function"))
	}

	if m == 0 {
		of.Printf("no functions matched for event %q on topic: %q", ce.ID(), topic)
		return nil
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// expect m callbacks
	for i := 0; i < m; i++ {
		res := <-of.respChan

		eg.Go(func() error {
			retry, err := isRetryable(egCtx, res.Status, res.Error)
			if err != nil {
				of.Printf("could not invoke function %q on topic %q: %v", res.Function, topic, err)
				// 	no retry
				return nil
			}

			if retry {
				of.Printf("function %q on topic %q returned non successful status code %d: %q", res.Function, res.Topic, res.Status, string(*res.Body))
				// 	TODO: retry logic
				return nil
			}

			of.Printf("successfully invoked function %q for topic %q", res.Function, res.Topic)
			return nil
		})

	}

	return eg.Wait()
}

// handleEvent returns the OpenFaaS subscription topic, e.g. VmPoweredOnEvent,
// and outbound event message ([]byte(CloudEvent) for the given CloudEvent
func handleEvent(event cloudevents.Event) (string, []byte, error) {
	message, err := json.Marshal(event)
	if err != nil {
		return "", nil, errors.Wrapf(err, "could not JSON-encode CloudEvent %v", event)
	}
	return event.Subject(), message, nil
}

func (of *OpenfaasProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			of.lock.RLock()
			ms.Receive(&of.stats)
			of.lock.RUnlock()
		}
	}
}

// isRetryable provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func isRetryable(ctx context.Context, code int, err error) (bool, error) {
	// source: https://github.com/hashicorp/go-retryablehttp/blob/master/client.go
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe := regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe := regexp.MustCompile(`unsupported protocol scheme`)

	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, nil
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if code == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if code == 0 || (code >= 500 && code != 501) {
		return true, nil
	}

	return false, nil
}
