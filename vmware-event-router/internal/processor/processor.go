package processor

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
)

// Processor handles incoming stream events to decouple event stream providers,
// e.g. vCenter, from processors, e.g. OpenFaaS, knative, AWS EventBridge, etc.
type Processor interface {
	Process(cloudevents.Event) error
}

// Error struct contains the generic error content used by the processors
// it extends the simple error by providing context which processor gave
// the error
type Error struct {
	processor config.ProcessorType
	err       error
}

func processorError(processor config.ProcessorType, err error) error {
	return &Error{
		processor: processor,
		err:       err,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.processor, e.err.Error())
}
