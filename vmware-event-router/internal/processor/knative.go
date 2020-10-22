package processor

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	knativeDefaultRetryInterval = time.Second * 3
	knativeDefaultMaxRetries    = 3
)

// knativeProcessor implements the Processor interface
type knativeProcessor struct {
	address       string
	verbose       bool
	retryInterval time.Duration
	maxReTries    int
	client        client.Client
	*log.Logger
	lock  sync.RWMutex
	stats metrics.EventStats
}

func NewKnativeProcessor(ctx context.Context, cfg *config.ProcessorConfigKnative, ms metrics.Receiver, opts ...KnativeOption) (Processor, error) {

	logger := log.New(os.Stdout, color.Purple("[Knative] "), log.LstdFlags)
	kProcessor := knativeProcessor{
		Logger:        logger,
		retryInterval: knativeDefaultRetryInterval,
		maxReTries:    knativeDefaultMaxRetries,
	}

	// apply options
	for _, opt := range opts {
		opt(&kProcessor)
	}

	if len(cfg.Address) > 0 {
		kProcessor.address = cfg.Address
	}

	c, err := cloudevents.NewDefaultClient()

	if err != nil {
		logger.Printf("failed to create knative client, %v\n", err)
		return nil, errors.New("failed to create knative client")
	}
	kProcessor.client = c

	if cfg == nil {
		return nil, errors.New("no Knative configuration found")
	}

	//Starting scheduler for updating knative Metrics.
	go kProcessor.PushMetrics(ctx, ms)

	return &kProcessor, nil

}

// Process implements the stream processor interface
func (kProcessor *knativeProcessor) Process(ce cloudevents.Event) error {

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), kProcessor.address)
	//Send events, retrying in case of a failure.
	ctx = cloudevents.ContextWithRetriesLinearBackoff(ctx, kProcessor.retryInterval, kProcessor.maxReTries)

	// Send that Event.
	if result := kProcessor.client.Send(ctx, ce); cloudevents.IsACK(result) {

		kProcessor.lock.Lock()
		total := *kProcessor.stats.EventsTotal + 1
		kProcessor.stats.EventsTotal = &total
		kProcessor.lock.Unlock()

		kProcessor.Printf("Sent: %s", ce.ID())
	} else if cloudevents.IsNACK(result) {

		kProcessor.lock.Lock()
		errTotal := *kProcessor.stats.EventsErr + 1
		kProcessor.stats.EventsErr = &errTotal
		kProcessor.lock.Unlock()

		kProcessor.Printf("Sent but not accepted: %s", result.Error())
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
