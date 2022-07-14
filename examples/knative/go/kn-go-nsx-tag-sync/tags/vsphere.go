package tags

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func getInstancedID(ctx context.Context, client *vim25.Client, ref types.ManagedObjectReference) (string, error) {
	var info mo.VirtualMachine
	pc := property.DefaultCollector(client)
	if err := pc.RetrieveOne(ctx, ref, []string{"config.instanceUuid"}, &info); err != nil {
		return "", fmt.Errorf("retrieve instanceUuid property: %w", err)
	}

	return info.Config.InstanceUuid, nil
}

func getVmRef(ctx context.Context, client *vim25.Client, name string) (types.ManagedObjectReference, error) {
	f := find.NewFinder(client)
	vm, err := f.VirtualMachine(ctx, name)
	if err != nil {
		return types.ManagedObjectReference{}, err
	}

	return vm.Reference(), nil
}
