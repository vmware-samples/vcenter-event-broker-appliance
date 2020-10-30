package processor

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	knativeDefaultRetryInterval = time.Second * 3
	knativeDefaultMaxRetries    = 3
)

// KnativeProcessor implements the Processor interface
type knativeProcessor struct {
	address string
	verbose bool
	client  cloudevents.Client
	log     *log.Logger
	lock    sync.RWMutex
	stats   metrics.EventStats
}

// NewKnativeProcessor method creates a Knative Processor
func NewKnativeProcessor(ctx context.Context, cfg *config.ProcessorConfigKnative, ms metrics.Receiver, opts ...KnativeOption) (Processor, error) {
	logger := log.New(os.Stdout, color.Purple("[Knative] "), log.LstdFlags)
	kProcessor := knativeProcessor{
		log: logger,
	}
	// apply options
	for _, opt := range opts {
		opt(&kProcessor)
	}
	if cfg == nil {
		return nil, errors.New("no Knative configuration found")
	}
	if len(cfg.Address) <= 0 {
		return nil, errors.New("Broker address can not be null")
	}
	kProcessor.address = cfg.Address
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSSL,
	}
	httpTransport := &http.Transport{TLSClientConfig: tlsConfig}
	// Create protocol and client
	p, err := cloudevents.NewHTTP(cloudevents.WithRoundTripper(httpTransport))
	if err != nil {
		logger.Printf("failed to create protocol for knative client, %v\n", err)
		return nil, errors.New("failed to create protocol for knative client")
	}
	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow())
	if err != nil {
		logger.Printf("failed to create knative client, %v\n", err)
		return nil, errors.New("failed to create knative client")
	}
	kProcessor.client = c
	//Starting scheduler for updating Knative Metrics.
	go kProcessor.PushMetrics(ctx, ms)
	return &kProcessor, nil
}

// Process implements the stream processor interface
func (kProcessor *knativeProcessor) Process(ce cloudevents.Event) error {
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), kProcessor.address)
	//Retries in case of a failure.
	ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, knativeDefaultRetryInterval, knativeDefaultMaxRetries)
	// Send that Event.
	if result := kProcessor.client.Send(ctx, ce); cloudevents.IsACK(result) {
		kProcessor.lock.Lock()
		total := *kProcessor.stats.EventsTotal + 1
		kProcessor.stats.EventsTotal = &total
		kProcessor.lock.Unlock()
		kProcessor.log.Printf("Sent: %s", ce.ID())
		return nil
	} else if cloudevents.IsNACK(result) { //Continue with error path
		kProcessor.lock.Lock()
		errTotal := *kProcessor.stats.EventsErr + 1
		kProcessor.stats.EventsErr = &errTotal
		kProcessor.lock.Unlock()
		kProcessor.log.Printf("Event ID %s Sent but not accepted: %s", ce.ID(), result.Error())
		if kProcessor.verbose {
			kProcessor.log.Printf("The result of the sending event to Knative broker is : %v", result)
		}
	}
	return nil
}

func (kProcessor *knativeProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			kProcessor.lock.RLock()
			ms.Receive(&kProcessor.stats)
			kProcessor.lock.RUnlock()
		}
	}
}
