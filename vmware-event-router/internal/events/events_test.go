// +build unit

package events

import (
	"testing"

	"github.com/vmware/govmomi/vim25/types"
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
			// TODO: handle error
			if got, _ := getDetails(tt.args.event); got != tt.want {
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
