// +build unit

package events

import (
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi/vim25/types"
	"gotest.tools/assert"
)

func Test_GetEventDetails(t *testing.T) {
	type args struct {
		event types.BaseEvent
	}
	tests := []struct {
		name string
		args args
		want VCenterEventInfo
	}{
		{
			name: "Event: VmPoweredOnEvent",
			args: args{newVMPoweredOnEvent()},
			want: VCenterEventInfo{Category: "event", Name: "VmPoweredOnEvent"},
		},
		{
			name: "EventEx: com.vmware.cl.PublishLibraryEvent",
			args: args{newEventExEvent()},
			want: VCenterEventInfo{Category: "eventex", Name: "com.vmware.cl.PublishLibraryEvent"},
		},
		{
			name: "ExtendedEvent: com.vmware.applmgmt.backup.job.failed.event",
			args: args{newExtendedEvent()},
			want: VCenterEventInfo{Category: "extendedevent", Name: "com.vmware.applmgmt.backup.job.failed.event"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDetails(tt.args.event); got != tt.want {
				t.Errorf("getEventDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newVMPoweredOnEvent() types.BaseEvent {
	return &types.VmPoweredOnEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Vm: &types.VmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: "Linux-1234",
					},
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1234",
					},
				},
			},
		},
	}
}

func newExtendedEvent() types.BaseEvent {
	return &types.ExtendedEvent{
		EventTypeId: "com.vmware.applmgmt.backup.job.failed.event",
	}
}

func newEventExEvent() types.BaseEvent {
	return &types.EventEx{
		EventTypeId: "com.vmware.cl.PublishLibraryEvent",
	}
}

func Test_NewFromVSphere(t *testing.T) {
	const (
		source = "https://vcenter.local/sdk"
	)

	now := time.Now().UTC()

	// event without extension attributes
	e1 := cloudevents.NewEvent()
	e1.SetSource(source)
	e1.SetID("1")
	e1.SetTime(now)
	e1.SetType(eventCanonicalType + "/" + "event")
	e1.SetSubject("VmPoweredOnEvent")
	if err := e1.SetData(cloudevents.ApplicationJSON, newVMPoweredOnEvent()); err != nil {
		t.Errorf("marshal data: %v", err)
	}

	e2 := e1.Clone()
	e2.SetExtension("vsphereapiversion", "6.7.3")

	testEvents := []cloudevents.Event{e1, e2}

	type args struct {
		event  types.BaseEvent
		source string
		opts   []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *cloudevents.Event
		wantErr bool
	}{
		{
			name: "event without extension attributes",
			args: args{
				event:  newVMPoweredOnEvent(),
				source: source,
				opts: []Option{
					WithTime(now),
					WithID("1"),
				},
			},
			want:    &testEvents[0],
			wantErr: false,
		},
		{
			name: "event extension attributes",
			args: args{
				event:  newVMPoweredOnEvent(),
				source: source,
				opts: []Option{
					WithTime(now),
					WithID("1"),
					WithAttributes(map[string]string{"vsphereapiversion": "6.7.3"}),
				},
			},
			want:    &testEvents[1],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromVSphere(tt.args.event, tt.args.source, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromVSphere() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.DeepEqual(t, tt.want, got)
		})
	}
}
