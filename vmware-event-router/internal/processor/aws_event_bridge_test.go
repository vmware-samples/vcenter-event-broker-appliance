package processor

import (
	"testing"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware/govmomi/vim25/types"
)

func Test_batching_createPutEventsInput(t *testing.T) {
	tests := []struct {
		title          string
		baseEvents     []types.BaseEvent
		desiredBatches int
		desiredEntries int
		desiredType    string
	}{
		{
			title:          "13 VmPoweredOnEvent events 2 batches",
			baseEvents:     baseEventsMockVMPoweredOn(13),
			desiredBatches: 2,
			desiredEntries: 13,
			desiredType:    "VmPoweredOnEvent",
		},
		{
			title:          "10 VmPoweredOnEvent events 1 batch",
			baseEvents:     baseEventsMockVMPoweredOn(10),
			desiredBatches: 1,
			desiredEntries: 10,
			desiredType:    "VmPoweredOnEvent",
		},
		{
			title:          "3 VmPoweredOnEvent events 1 batch",
			baseEvents:     baseEventsMockVMPoweredOn(3),
			desiredBatches: 1,
			desiredEntries: 3,
			desiredType:    "VmPoweredOnEvent",
		},
		{
			title:          "23 VmPoweredOnEvent events 3 batches",
			baseEvents:     baseEventsMockVMPoweredOn(23),
			desiredBatches: 3,
			desiredEntries: 23,
			desiredType:    "VmPoweredOnEvent",
		},
		{
			title:          "0 events 0 batches unsubscribed event",
			baseEvents:     baseEventsMockCustomizedDVPortEvent(23),
			desiredBatches: 0,
			desiredEntries: 0,
			desiredType:    "",
		},
		{
			title:          "5 VmPoweredOnEvent events 1 batches",
			baseEvents:     baseEventsMockHalfVMPoweredOn(10),
			desiredBatches: 1,
			desiredEntries: 5,
			desiredType:    "VmPoweredOnEvent",
		},
		{
			title:          "0 events 0 batches",
			baseEvents:     []types.BaseEvent{},
			desiredBatches: 0,
			desiredEntries: 0,
			desiredType:    "",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			awsEventBridgeStub := createAWSObjectStubVMPoweredOn()
			batchEvents, err := awsEventBridgeStub.createPutEventsInput(test.baseEvents)
			if err != nil {
				t.Errorf("unexpected error: %s", err.Error())
			}
			actualBatches := 0
			actualEntries := 0
			for _, v := range batchEvents {
				for _, entry := range v.Entries {
					actualEntries++
					if test.desiredType != *entry.DetailType {
						t.Errorf("wanted entry type: %s got: %s",
							test.desiredType,
							*entry.DetailType)
					}
				}
				actualBatches++
			}
			if test.desiredBatches != actualBatches {
				t.Errorf("wanted: %v batches got: %v",
					test.desiredBatches,
					actualBatches)
			}
			if test.desiredEntries != actualEntries {
				t.Errorf("wanted: %v entries got: %v",
					test.desiredEntries,
					actualEntries)
			}
		})
	}
}

func baseEventsMockVMPoweredOn(numberOfEvents int) []types.BaseEvent {
	baseEvents := []types.BaseEvent{}
	for numberOfEvents > 0 {
		numberOfEvents = numberOfEvents - 1
		baseEvents = append(baseEvents, &types.VmPoweredOnEvent{})
	}
	return baseEvents
}

func baseEventsMockCustomizedDVPortEvent(numberOfEvents int) []types.BaseEvent {
	baseEvents := []types.BaseEvent{}
	for numberOfEvents > 0 {
		numberOfEvents = numberOfEvents - 1
		baseEvents = append(baseEvents, &types.VmPoweringOnWithCustomizedDVPortEvent{})
	}
	return baseEvents
}

func baseEventsMockHalfVMPoweredOn(numberOfEvents int) []types.BaseEvent {
	baseEvents := []types.BaseEvent{}
	switchFlag := true
	for numberOfEvents > 0 {
		if switchFlag {
			baseEvents = append(baseEvents, &types.VmPoweringOnWithCustomizedDVPortEvent{})
			switchFlag = false
		} else {
			baseEvents = append(baseEvents, &types.VmPoweredOnEvent{})
			switchFlag = true
		}
		numberOfEvents = numberOfEvents - 1
	}
	return baseEvents
}

func createAWSObjectStubVMPoweredOn() awsEventBridgeProcessor {
	return awsEventBridgeProcessor{
		patternMap: map[string]string{"VmPoweredOnEvent": ""},
		stats: metrics.EventStats{
			Invocations: make(map[string]int),
		},
	}
}
