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
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	// ProviderOpenFaaS variable is the name used to identify this provider in
	// the VMware Event Router configuration file
	ProviderOpenFaaS       = "openfaas"
	authMethodOpenFaaS     = "basic_auth" // only this method is supported by the processor
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

// openfaasProcessor implements the Processor interface
type openfaasProcessor struct {
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
// source. Asynchronous function invokation can be configured for
// high-throughput (non-blocking) requirements.
func NewOpenFaaSProcessor(ctx context.Context, cfg connection.Config, ms metrics.Receiver, opts ...OpenFaaSOption) (Processor, error) {
	// defaults
	logger := log.New(os.Stdout, color.Purple("[OpenFaaS] "), log.LstdFlags)
	ofProcessor := openfaasProcessor{
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

	var creds auth.BasicAuthCredentials
	switch cfg.Auth.Method {
	case authMethodOpenFaaS:
		creds.User = cfg.Auth.Secret["username"]
		creds.Password = cfg.Auth.Secret["password"]
	default:
		return nil, errors.Errorf("unsupported authentication method for processor openfaas: %s", cfg.Auth.Method)
	}

	var async bool
	if cfg.Options["async"] == "true" {
		async = true
	}
	ofconfig := ofsdk.ControllerConfig{
		GatewayURL:               cfg.Address,
		TopicAnnotationDelimiter: ofProcessor.topicDelimiter,
		RebuildInterval:          ofProcessor.rebuildInterval,
		UpstreamTimeout:          ofProcessor.gatewayTimeout,
		AsyncFunctionInvocation:  async,
		PrintSync:                ofProcessor.verbose,
	}
	ofcontroller := ofsdk.NewController(&creds, &ofconfig)
	ofProcessor.controller = ofcontroller
	ofProcessor.controller.Subscribe(&ofProcessor)
	ofProcessor.controller.BeginMapBuilder()

	// prepopulate the metrics stats
	ofProcessor.stats = metrics.EventStats{
		Provider:     ProviderOpenFaaS,
		ProviderType: cfg.Type,
		Name:         cfg.Address,
		Started:      time.Now().UTC(),
		Invocations:  make(map[string]int),
	}
	go ofProcessor.PushMetrics(ctx, ms)

	return &ofProcessor, nil
}

// defaultResponseHandler prints status information for each function invokation
func defaultResponseHandler(openfaas *openfaasProcessor) responseFunc {
	return func(res ofsdk.InvokerResponse) {
		// update stats
		// TODO: currently we only support metrics when in sync invokation mode
		// because we don't have a callback for async invocations
		openfaas.lock.Lock()
		openfaas.stats.Invocations[res.Topic]++
		openfaas.lock.Unlock()

		if res.Error != nil || res.Status != http.StatusOK {
			openfaas.Printf("function %s for topic %s returned status %d with error: %v", res.Function, res.Topic, res.Status, res.Error)
			return
		}
		openfaas.Printf("successfully invoked function %s for topic %s", res.Function, res.Topic)
	}
}

// Process implements the stream processor interface
func (openfaas *openfaasProcessor) Process(ce cloudevents.Event) error {
	if openfaas.verbose {
		openfaas.Printf("processing event (ID %s): %v", ce.ID(), ce)
	}

	topic, message, err := handleEvent(ce)
	if err != nil {
		msg := fmt.Errorf("error handling event %v: %v", ce, err)
		openfaas.Println(msg)
		return processorError(ProviderOpenFaaS, msg)
	}

	if openfaas.verbose {
		openfaas.Printf("created new outbound event for subscribers: %s", string(message))
	}

	openfaas.Printf("invoking function(s) for event %s on topic: %s", ce.ID(), topic)
	openfaas.controller.Invoke(topic, &message)
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

func (openfaas *openfaasProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			openfaas.lock.RLock()
			ms.Receive(openfaas.stats)
			openfaas.lock.RUnlock()
		}
	}
}
