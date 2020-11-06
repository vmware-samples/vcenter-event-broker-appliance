// +build unit

package vcsim

import (
	"context"
	"reflect"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap/zaptest"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

func Test_reverse(t *testing.T) {
	vmOnEvent := types.VmPoweredOnEvent{}
	vmOffEvent := types.VmPoweredOffEvent{}

	type args struct {
		events []types.BaseEvent
	}
	tests := []struct {
		name string
		args args
		want []types.BaseEvent
	}{
		{
			name: "empty slice",
			args: args{
				events: nil,
			},
			want: nil,
		},
		{
			name: "two entries",
			args: args{
				events: []types.BaseEvent{&vmOnEvent, &vmOffEvent},
			},
			want: []types.BaseEvent{&vmOffEvent, &vmOnEvent},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reverse(tt.args.events)
			if !reflect.DeepEqual(tt.args.events, tt.want) {
				t.Errorf("reverse() gotReverse = %v, want %v", tt.args.events, tt.want)
			}
		})
	}
}

func Test_eventHandler(t *testing.T) {
	type eventStats struct {
		total  int
		errors int
	}

	vmOnEvent := types.VmPoweredOnEvent{}
	vmOffEvent := types.VmPoweredOffEvent{}

	u, err := soap.ParseURL("https://127.0.0.1:8989/sdk")
	if err != nil {
		t.Fatalf("could not parse url: %v", err)
	}
	c := soap.NewClient(u, false)

	zero := 0
	fakeSim := EventStream{
		client: govmomi.Client{
			Client: &vim25.Client{
				Client: c,
			},
		},
		Logger: zaptest.NewLogger(t).Sugar(),
		stats: metrics.EventStats{
			EventsTotal: &zero,
			EventsErr:   &zero,
		},
	}

	type args struct {
		ctx   context.Context
		vcsim *EventStream
		proc  processor.Processor
	}
	tests := []struct {
		name    string
		args    args
		want    eventStats
		wantErr bool
	}{
		{
			name: "Two events, one not successful",
			args: args{
				ctx:   context.Background(),
				vcsim: &fakeSim,
				proc:  fakeProcessor{},
			},
			want: eventStats{
				total:  2,
				errors: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eh := eventHandler(tt.args.ctx, tt.args.vcsim, tt.args.proc)
			if err := eh(types.ManagedObjectReference{}, []types.BaseEvent{&vmOnEvent, &vmOffEvent}); (err != nil) != tt.wantErr {
				t.Errorf("eventHandler() = %v, wantErr %v", err, tt.wantErr)
			}

			fakeSim.Lock()
			total := *fakeSim.stats.EventsTotal
			errs := *fakeSim.stats.EventsErr
			fakeSim.Unlock()

			if total != tt.want.total {
				t.Errorf("eventHandler() = %d, wantTotal %d", total, tt.want.total)
			}

			if errs != tt.want.errors {
				t.Errorf("eventHandler() = %d, wantTotal %d", errs, tt.want.errors)
			}
		})
	}
}

// fakeProcessor implements the processor interface
type fakeProcessor struct {
}

func (f fakeProcessor) PushMetrics(_ context.Context, _ metrics.Receiver) {
	return
}

func (f fakeProcessor) Shutdown(_ context.Context) error {
	return nil
}

// Process processes an event and returns an error if it is type
// VmPoweredOnEvent
func (f fakeProcessor) Process(_ context.Context, event cloudevents.Event) error {
	switch event.Subject() {
	case "VmPoweredOnEvent":
		return errors.New("failed")
	default:
		return nil
	}
}
