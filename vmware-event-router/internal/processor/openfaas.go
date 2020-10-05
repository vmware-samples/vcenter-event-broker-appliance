package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"

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

// defaultResponseHandler prints status information for each function invocation
func defaultResponseHandler(of *OpenfaasProcessor) responseFunc {
	return func(res ofsdk.InvokerResponse) {
		// update stats
		// TODO: currently we only support metrics when in sync invocation mode
		// because we don't have a callback for async invocations
		of.lock.Lock()
		of.stats.Invocations[res.Topic]++
		of.lock.Unlock()

		if res.Error != nil || res.Status != http.StatusOK {
			of.Printf("function %s for topic %s returned status %d with error: %v", res.Function, res.Topic, res.Status, res.Error)
			return
		}

		of.Printf("successfully invoked function %s for topic %s", res.Function, res.Topic)
	}
}

// Process implements the stream processor interface
func (of *OpenfaasProcessor) Process(ce cloudevents.Event) error {
	if of.verbose {
		of.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	topic, message, err := handleEvent(ce)
	if err != nil {
		msg := fmt.Errorf("error handling event %v: %v", ce, err)
		of.Println(msg)
		return processorError(config.ProcessorOpenFaaS, msg)
	}

	if of.verbose {
		of.Printf("created new outbound event for subscribers: %s", string(message))
	}

	of.Printf("invoking function(s) for event %s on topic: %s", ce.ID(), topic)
	of.controller.Invoke(topic, &message)
	return nil
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
