package stream

import (
	"context"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// ProviderVSphere is the name used to identify this provider in the
	// VMware Event Router configuration file
	ProviderVSphere   = "vmware_vcenter"
	authMethodvSphere = "user_password"
)

// vCenterStream handles the connection to vCenterStream to retrieve an event stream
type vCenterStream struct {
	client govmomi.Client
	stream event.Manager

	lock  sync.RWMutex
	stats metrics.EventStats
}

// NewVCenterStream returns a vCenter event manager for a given configuration and metrics server
func NewVCenterStream(ctx context.Context, cfg connection.Config, ms *metrics.Server) (Streamer, error) {
	var vCenter vCenterStream
	parsedURL, err := soap.ParseURL(cfg.Address)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing URL")
	}

	var username, password string
	switch cfg.Auth.Method {
	case authMethodvSphere:
		username = cfg.Auth.Secret["username"]
		password = cfg.Auth.Secret["password"]
	default:
		return nil, errors.Errorf("unsupported authentication method for stream vCenter: %s", cfg.Auth.Method)
	}
	parsedURL.User = url.UserPassword(username, password)

	var insecure bool
	if cfg.Options["insecure"] == "true" {
		insecure = true
	}

	client, err := govmomi.NewClient(ctx, parsedURL, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "could not create vCenter client")
	}

	vCenter.client = *client
	vCenter.stream = *event.NewManager(client.Client)

	// prepopulate the metrics stats
	vCenter.stats = metrics.EventStats{
		Provider:     ProviderVSphere,
		ProviderType: cfg.Type,
		Name:         client.URL().String(),
		Started:      time.Now().UTC(),
		EventsTotal:  new(int),
		EventsSec:    new(float64),
	}
	go vCenter.PushMetrics(ctx, ms)
	return &vCenter, nil
}

// Stream is the main logic, blocking to receive and handle events from vCenter
func (vcenter *vCenterStream) Stream(ctx context.Context, p processor.Processor) error {
	// get events for all types (i.e. RootFolder in vCenter)
	managedTypes := []types.ManagedObjectReference{vcenter.client.ServiceContent.RootFolder}
	eventsPerPage := int32(1)
	tail := true
	force := false

	err := vcenter.stream.Events(ctx, managedTypes, eventsPerPage, tail, force, vcenter.streamCallbackFn(p))
	if err != nil {
		return errors.Wrap(err, "error connecting to vCenter event stream")
	}
	return nil
}

func (vcenter *vCenterStream) Shutdown(ctx context.Context) error {
	// need to pass new context explicitly to avoid
	// "*url.Error: POST ... context cancelled"
	err := vcenter.client.Logout(context.Background())
	return errors.Wrap(err, "failed to logout from vCenter") // err == nil if logout was successful
}

func (vcenter *vCenterStream) Source() string {
	return vcenter.client.URL().String()
}

func (vcenter *vCenterStream) PushMetrics(ctx context.Context, ms *metrics.Server) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vcenter.lock.RLock()
			eventsSec := math.Round((float64(*vcenter.stats.EventsTotal)/time.Since(vcenter.stats.Started).Seconds())*100) / 100 // 0.2f syntax
			vcenter.stats.EventsSec = &eventsSec
			ms.Receive(vcenter.stats)
			vcenter.lock.RUnlock()
		}
	}
}

// updates the internal metrics state of the provider before invoking the
// processor
func (vcenter *vCenterStream) streamCallbackFn(p processor.Processor) func(types.ManagedObjectReference, []types.BaseEvent) error {
	return func(moref types.ManagedObjectReference, baseEvent []types.BaseEvent) error {
		// update stats before invoking the processor
		vcenter.lock.Lock()
		total := *vcenter.stats.EventsTotal + len(baseEvent)
		vcenter.stats.EventsTotal = &total
		vcenter.lock.Unlock()

		return p.Process(moref, baseEvent)
	}
}
