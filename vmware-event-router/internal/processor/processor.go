package processor

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/types"
)

// Processor handles incoming vCenter events. This enables different FaaS
// implementations for vCenter event processing. Note: in the case of processing
// failure the current behavior is to log but return nil until at-least-once
// semantics are implemented.
type Processor interface {
	Process(types.ManagedObjectReference, []types.BaseEvent) error
}

// Error struct contains the generic error content used by the processors
// it extends the simple error by providing context which processor gave
// the error
type Error struct {
	processor string
	err       error
}

// NewError constructs Error structure
func NewError(processor string, err error) *Error {
	return &Error{
		processor: processor,
		err:       err,
	}
}

func (e *Error) Error() error { return fmt.Errorf("%s: %s", e.processor, e.err.Error()) }
