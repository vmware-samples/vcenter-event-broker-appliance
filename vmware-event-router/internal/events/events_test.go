package events

import (
	"encoding/json"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
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
			if got := GetDetails(tt.args.event); got != tt.want {
				t.Errorf("getEventDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConvertToCloudEventV1(t *testing.T) {
	vmEvent := newVMPoweredOnEvent()
	eInfo := GetDetails(vmEvent)
	e := NewCloudEvent(vmEvent, eInfo, getSource())
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("could not marshal cloud event: %v", err)
	}

	ce := cloudevents.NewEvent(cloudevents.VersionV1)
	err = ce.UnmarshalJSON(b)
	if err != nil {
		t.Fatalf("could not unmarshal outbound cloud event into cloud events v1 spec: %v", err)
	}

	if e.ID != ce.ID() {
		t.Fatalf("ID of outbound cloud event and cloud event v1 does not match: %q vs %q", e.ID, ce.ID())
	}

	if e.Source != ce.Source() {
		t.Fatalf("Source of outbound cloud event and cloud event v1 does not match: %q vs %q", e.Source, ce.Source())
	}

	if e.SpecVersion != ce.SpecVersion() {
		t.Fatalf("SpecVersions of outbound cloud event and cloud event v1 does not match: %q vs %q", e.SpecVersion, ce.SpecVersion())
	}

	if e.Subject != ce.Subject() {
		t.Fatalf("Subject of outbound cloud event and cloud event v1 don't match: %q vs %q", e.Subject, ce.Subject())
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

func getSource() string {
	return "https://vcenter.corp.local/sdk"
}
