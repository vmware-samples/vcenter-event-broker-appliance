package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	sdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"

	"github.com/vmware/govmomi/vim25/types"
)

const (
	// ProviderOpenFaaS is the name used to identify this provider in the
	// VMware Event Router configuration file
	ProviderOpenFaaS   = "openfaas"
	topicDelimiter     = ","
	rebuildInterval    = time.Second * 10
	timeout            = time.Second * 15
	authMethodOpenFaaS = "basic_auth" // only this method is supported by the processor
)

// openfaasProcessor implements the Processor interface
type openfaasProcessor struct {
	controller sdk.Controller
	source     string
	verbose    bool
	*log.Logger

	lock  sync.RWMutex
	stats metrics.EventStats
}

// NewOpenFaaSProcessor returns an OpenFaaS processor for the given stream
// source. Asynchronous function invokation can be configured for
// high-throughput (non-blocking) requirements.
func NewOpenFaaSProcessor(ctx context.Context, cfg connection.Config, source string, verbose bool, ms *metrics.Server) (Processor, error) {
	logger := log.New(os.Stdout, "[OpenFaaS] ", log.LstdFlags)
	openfaas := openfaasProcessor{
		source:  source,
		verbose: verbose,
		Logger:  logger,
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
	ofconfig := sdk.ControllerConfig{
		GatewayURL:               cfg.Address,
		TopicAnnotationDelimiter: topicDelimiter,
		RebuildInterval:          rebuildInterval,
		UpstreamTimeout:          timeout,
		AsyncFunctionInvocation:  async,
		PrintSync:                verbose,
	}
	ofcontroller := sdk.NewController(&creds, &ofconfig)

	openfaas.controller = ofcontroller
	openfaas.controller.Subscribe(&openfaas)
	openfaas.controller.BeginMapBuilder()

	// prepopulate the metrics stats
	openfaas.stats = metrics.EventStats{
		Provider:     ProviderOpenFaaS,
		ProviderType: cfg.Type,
		Name:         cfg.Address,
		Started:      time.Now().UTC(),
		Invocations:  make(map[string]int),
	}
	go openfaas.PushMetrics(ctx, ms)

	return &openfaas, nil
}

// Response prints status information for each function invokation
func (openfaas *openfaasProcessor) Response(res sdk.InvokerResponse) {
	// update stats 
	// TODO: currently we only support metrics when in sync invokation mode
	// because we don't have a callback for async invocations
	openfaas.lock.Lock()
	openfaas.stats.Invocations[res.Topic]++
	openfaas.lock.Unlock()

	if res.Error != nil {
		openfaas.Printf("function %s for topic %s returned status %d with error: %v", res.Function, res.Topic, res.Status, res.Error)
		return
	}
	openfaas.Printf("successfully invoked function %s for topic %s", res.Function, res.Topic)
}

// Process implements the stream processor interface
func (openfaas *openfaasProcessor) Process(moref types.ManagedObjectReference, baseEvent []types.BaseEvent) error {
	fmt.Printf("of topics: %v", openfaas.controller.Topics())

	for idx := range baseEvent {
		// process slice in reverse order to maintain Event.Key ordering
		event := baseEvent[len(baseEvent)-1-idx]

		if openfaas.verbose {
			openfaas.Printf("processing event [%d] of type %T from source %s: %+v", idx, event, openfaas.source, event)
		}

		topic, message, err := handleEvent(event, openfaas.source)
		if err != nil {
			openfaas.Printf("error handling event: %v", err)
			continue
		}

		if openfaas.verbose {
			openfaas.Printf("created new outbound cloud event for subscribers: %s", string(message))
		}

		openfaas.Printf("invoking function(s) on topic: %s", topic)
		openfaas.controller.Invoke(topic, &message)
	}
	return nil
}

// handleEvent returns the OpenFaaS subscription topic, e.g. VmPoweredOnEvent,
// and outbound event message for the given BaseEvent and source
func handleEvent(event types.BaseEvent, source string) (string, []byte, error) {
	// Sanity check to avoid nil pointer exception
	if event == nil {
		return "", nil, errors.New("source event must not be nil")
	}

	// Get the category and name of the event used for subscribed topic matching
	eventInfo := events.GetDetails(event)
	cloudEvent := events.NewCloudEvent(event, eventInfo, source)
	message, err := json.Marshal(cloudEvent)
	if err != nil {
		return "", nil, errors.Wrapf(err, "could not marshal cloud event for vSphere event %s from source %s", event.GetEvent().Key, source)
	}
	return eventInfo.Name, message, nil
}

func (openfaas *openfaasProcessor) PushMetrics(ctx context.Context, ms *metrics.Server) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(metrics.PushInterval):
			openfaas.lock.RLock()
			ms.Receive(openfaas.stats)
			openfaas.lock.RUnlock()
		}
	}
}
