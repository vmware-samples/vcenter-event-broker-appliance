//go:build integration

package integration_test

import "github.com/vmware/govmomi/vim25/types"

//nolint
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

//nolint
func newLicenseEvent() types.BaseEvent {
	return types.BaseEvent(&types.LicenseEvent{})
}

//nolint
func newClusterCreatedEvent() types.BaseEvent {
	return types.BaseEvent(&types.ClusterCreatedEvent{})
}
