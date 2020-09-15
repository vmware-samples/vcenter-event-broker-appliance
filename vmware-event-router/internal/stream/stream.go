package stream

import (
	"context"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

// Streamer establishes a connection to a stream provider and invokes a stream
// processor.
type Streamer interface {
	PushMetrics(context.Context, metrics.Receiver)
	Stream(context.Context, processor.Processor) error
	Shutdown(context.Context) error
}
