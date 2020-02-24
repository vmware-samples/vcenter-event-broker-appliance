package processor

import (
	"github.com/vmware/govmomi/vim25/types"
)

// Processor handles incoming vCenter events. This enables different FaaS
// implementations for vCenter event processing. Note: in the case of processing
// failure the current behavior is to log but return nil until at-least-once
// semantics are implemented.
type Processor interface {
	Process(types.ManagedObjectReference, []types.BaseEvent) error
}
