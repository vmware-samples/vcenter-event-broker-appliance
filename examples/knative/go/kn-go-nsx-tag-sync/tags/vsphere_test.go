package tags

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"gotest.tools/v3/assert"
)

func Test_getVmRef(t *testing.T) {
	tests := []struct {
		name    string
		vm      string
		want    types.ManagedObjectReference
		wantErr string
	}{
		{
			name: "retrieves ref for vm",
			vm:   "DC0_H0_VM0",
			want: types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "vm-55",
			},
			wantErr: "",
		},
		{
			name: "vm does not exist",
			vm:   "vm-not-exist",
			want: types.ManagedObjectReference{
				Type:  "",
				Value: "",
			},
			wantErr: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simulator.Run(func(ctx context.Context, client *vim25.Client) error {
				ref, err := getVmRef(ctx, client, tt.vm)
				assert.Equal(t, tt.want, ref)

				if tt.wantErr != "" {
					assert.ErrorContains(t, err, tt.wantErr)
				} else {
					assert.NilError(t, err)
				}

				return nil
			})
		})
	}
}

func Test_getInstancedID(t *testing.T) {
	tests := []struct {
		name    string
		vm      types.ManagedObjectReference
		wantErr string
	}{
		{
			name: "retrieves id for vm",
			vm: types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "vm-55",
			},
			wantErr: "",
		},
		{
			name: "vm does not exist",
			vm: types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "invalid",
			},
			wantErr: "has already been deleted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simulator.Run(func(ctx context.Context, client *vim25.Client) error {
				id, err := getInstancedID(ctx, client, tt.vm)

				if tt.wantErr != "" {
					assert.Equal(t, id, "")
					assert.ErrorContains(t, err, tt.wantErr)
				} else {
					assert.NilError(t, err)
					assert.Assert(t, len(id) > 0)
				}

				return nil
			})
		})
	}
}
