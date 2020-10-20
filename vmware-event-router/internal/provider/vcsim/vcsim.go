package vcsim

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider"
)

// EventStream handles the connection to the vCenter events API
type EventStream struct {
	client govmomi.Client
	*log.Logger
	verbose bool

	sync.Mutex
	stats metrics.EventStats
}

// eventHandlerFunc is a callback passed to the event manager
type eventHandlerFunc func(moRef types.ManagedObjectReference, baseEvents []types.BaseEvent) error

// assert we implement the provider interface
var _ provider.Provider = (*EventStream)(nil)

// NewEventStream returns a vCenter simulator event stream manager for a given
// configuration and metrics server
func NewEventStream(ctx context.Context, cfg *config.ProviderConfigVCSIM, ms metrics.Receiver, opts ...Option) (*EventStream, error) {
	if cfg == nil {
		return nil, errors.New("vCenter simulator configuration must be provided")
	}

	var vcsim EventStream

	parsedURL, err := soap.ParseURL(cfg.Address)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing vCenter simulator URL")
	}

	// TODO: only supporting basic auth against vCenter simulator for now
	if cfg.Auth == nil || cfg.Auth.BasicAuth == nil {
		return nil, fmt.Errorf("invalid %s credentials: username and password must be set", config.BasicAuth)
	}

	username := cfg.Auth.BasicAuth.Username
	password := cfg.Auth.BasicAuth.Password
	parsedURL.User = url.UserPassword(username, password)

	client, err := govmomi.NewClient(ctx, parsedURL, cfg.InsecureSSL)
	if err != nil {
		return nil, errors.Wrap(err, "create vCenter simulator client")
	}

	l := log.New(os.Stdout, color.Fata("[VCSIM] "), log.LstdFlags)
	vcsim.Logger = l
	vcsim.client = *client

	// apply options (overwrite any defaults)
	for _, opt := range opts {
		opt(&vcsim)
	}

	// seed the metrics stats
	vcsim.stats = metrics.EventStats{
		Provider:    string(config.ProviderVCSIM),
		Type:        config.EventProvider,
		Address:     cfg.Address,
		Started:     time.Now().UTC(),
		EventsTotal: new(int),
		EventsErr:   new(int),
		EventsSec:   new(float64),
	}

	go vcsim.PushMetrics(ctx, ms)
	return &vcsim, nil
}

// Stream implements the event provider interface and starts the event stream
func (vcsim *EventStream) Stream(ctx context.Context, processor processor.Processor) error {
	mgr := event.NewManager(vcsim.client.Client)
	defer func() {
		// ignore error against vcsim
		_, _ = mgr.Destroy(ctx)
	}()

	const (
		pageSize = 10
		tail     = true
		force    = true
	)

	// get events for all objects
	ref := vcsim.client.ServiceContent.RootFolder
	handler := eventHandler(ctx, vcsim, processor)

	// blocks
	return mgr.Events(ctx, []types.ManagedObjectReference{ref}, pageSize, tail, force, handler)
}

func eventHandler(ctx context.Context, vcsim *EventStream, proc processor.Processor) eventHandlerFunc {
	var (
		errCount int
		source   = vcsim.client.URL().String()
	)

	return func(_ types.ManagedObjectReference, baseEvents []types.BaseEvent) error {
		if len(baseEvents) == 0 {
			return nil
		}

		// reverse slice because vcsim sends events in descending key order
		reverse(baseEvents)
		for _, e := range baseEvents {
			ce, err := events.NewCloudEvent(e, source)
			if err != nil {
				vcsim.Printf("skipping event %v because it could not be converted to CloudEvent format: %v", e, err)
				errCount++
				continue
			}

			err = proc.Process(ctx, *ce)
			if err != nil {
				vcsim.Printf("could not process event %v: %v", ce, err)
				errCount++
				continue
			}
		}

		// update metrics
		vcsim.Lock()
		total := *vcsim.stats.EventsTotal + len(baseEvents)
		vcsim.stats.EventsTotal = &total
		errTotal := *vcsim.stats.EventsErr + errCount
		vcsim.stats.EventsErr = &errTotal
		vcsim.Unlock()

		return nil
	}
}

// reverse reverses the order of the given slice
func reverse(events []types.BaseEvent) {
	for i := len(events)/2 - 1; i >= 0; i-- {
		opp := len(events) - 1 - i
		events[i], events[opp] = events[opp], events[i]
	}
}

// Shutdown closes the underlying connection to vCenter simulator
func (vcsim *EventStream) Shutdown(_ context.Context) error {
	// EventManager:EventManager does not implement: Destroy_Task
	vcsim.Println("provider shutdown successful")
	return nil
}

// PushMetrics pushes metrics to the configured metrics receiver
func (vcsim *EventStream) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vcsim.Lock()
			eventsSec := math.Round((float64(*vcsim.stats.EventsTotal)/time.Since(vcsim.stats.Started).Seconds())*100) / 100 // 0.2f syntax
			vcsim.stats.EventsSec = &eventsSec
			ms.Receive(&vcsim.stats)
			vcsim.Unlock()
		}
	}
}
