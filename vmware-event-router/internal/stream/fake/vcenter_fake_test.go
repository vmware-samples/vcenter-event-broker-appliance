// +build unit

package fake

import (
	"context"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi/vim25/types"
)

type noOpProcessor struct {
	invokations int
}

func (n *noOpProcessor) Process(event cloudevents.Event) error {
	n.invokations++
	return nil
}

func TestFakeVCenter_Stream(t *testing.T) {
	type fields struct {
		ctxTimeout time.Duration       // used to shutdown the stream
		genDelay   *time.Duration      // delay for generating events
		events     [][]types.BaseEvent // events to send and expected to receive
	}
	type args struct {
		p *noOpProcessor
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantEvents int
		wantErr    bool
	}{
		{
			name: "receive 8 events with context timeout of 200ms",
			fields: fields{
				ctxTimeout: 200 * time.Millisecond,
				genDelay:   makeDelay(50 * time.Millisecond),
				events:     [][]types.BaseEvent{createVMPoweredOnEvents(3), createVMPoweredOnEvents(5)},
			},
			wantEvents: 8,
			args: args{
				&noOpProcessor{},
			},
		},
		{
			name: "expect no events with context timeout of 200ms",
			fields: fields{
				ctxTimeout: 200 * time.Millisecond,
				genDelay:   makeDelay(50 * time.Millisecond),
				events:     nil,
			},
			wantEvents: 0,
			args: args{
				&noOpProcessor{},
			},
		},
		{
			name: "expect no events since no delay is specified",
			fields: fields{
				ctxTimeout: 200 * time.Millisecond,
				genDelay:   nil,
				events:     [][]types.BaseEvent{},
			},
			wantEvents: 0,
			args: args{
				&noOpProcessor{},
			},
		},
		{
			name: "expect no events since no events are specified",
			fields: fields{
				ctxTimeout: 200 * time.Millisecond,
				genDelay:   makeDelay(50 * time.Millisecond),
				events:     nil,
			},
			args: args{
				&noOpProcessor{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.fields.ctxTimeout)
			defer cancel()

			eventCh := createGenerator(ctx, tt.fields.genDelay, tt.fields.events)
			f := NewFakeVCenter(eventCh)
			_ = f.Stream(ctx, tt.args.p) // ignore errors

			got := tt.args.p.invokations
			want := tt.wantEvents
			if got != want {
				t.Errorf("FakeVCenter.Stream() invokations = %d, wanted %d", got, want)
			}
		})
	}
}

// the returned generator will send each []types.BaseEvent in the
// [][]types.BaseEvent slice via the returned channel. If neither a delay for
// the time between sending these events is specified or no events are specified
// at all a nil channel is returned which blocks forever
func createGenerator(ctx context.Context, delay *time.Duration, events [][]types.BaseEvent) <-chan []types.BaseEvent {
	if events == nil || delay == nil {
		return nil
	}

	genCh := make(chan []types.BaseEvent, len(events))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(*delay):
				// if we consumed all events set generator channel to block
				// forever and return
				if len(events) == 0 {
					genCh = nil
					return
				}

				baseEvent := events[0]
				genCh <- baseEvent
				events = events[1:]
			}
		}
	}()

	return genCh
}

func createVMPoweredOnEvents(count int) []types.BaseEvent {
	if count == 0 {
		return nil
	}

	var events []types.BaseEvent
	for i := 0; i < count; i++ {
		e := createVMPowerOnEvent()
		events = append(events, &e)
	}
	return events
}

func createVMPowerOnEvent() types.VmPoweredOnEvent {
	return types.VmPoweredOnEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Vm: &types.VmEventArgument{
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1234",
					},
				},
			},
		},
	}
}

func makeDelay(t time.Duration) *time.Duration {
	return &t
}
