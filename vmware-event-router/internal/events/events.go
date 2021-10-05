package events

import (
	"reflect"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// EventCanonicalType is the prefix used in the CloudEvent type by the VMware
	// Event Router
	EventCanonicalType = "com.vmware.event.router"
	// EventSpecVersion is the CloudEvent spec version used by the VMware Event
	// Router
	EventSpecVersion = cloudevents.VersionV1
	// EventContentType is the CloudEvent data content type used by the VMware Event
	// Router
	EventContentType = cloudevents.ApplicationJSON
)

// VCenterEventInfo contains the name and category of an event received from vCenter
// supported event categories: event, eventex, extendedevent
// category to name convention:
// event: retrieved from event class, e.g. VmPoweredOnEvent
// eventex: retrieved from EventTypeId
// extendedevent: retrieved from EventTypeId
type VCenterEventInfo struct {
	Category string
	Name     string
}

// GetDetails retrieves the underlying vSphere event category and name for
// the given BaseEvent, e.g. VmPoweredOnEvent (event) or
// com.vmware.applmgmt.backup.job.failed.event (extendedevent)
func GetDetails(event types.BaseEvent) VCenterEventInfo {
	var eventInfo VCenterEventInfo

	switch e := event.(type) {
	case *types.EventEx:
		eventInfo.Category = "eventex"
		eventInfo.Name = e.EventTypeId
	case *types.ExtendedEvent:
		eventInfo.Category = "extendedevent"
		eventInfo.Name = e.EventTypeId

	// TODO: make agnostic to vCenter events
	default:
		eType := reflect.TypeOf(event).Elem().Name()
		eventInfo.Category = "event"
		eventInfo.Name = eType
	}

	return eventInfo
}

type Option func(e *cloudevents.Event) error

// WithTime sets the provided time in the cloud event context
func WithTime(t time.Time) Option {
	return func(e *cloudevents.Event) error {
		e.SetTime(t)
		return nil
	}
}

// WithID sets the provided ID in the cloud event context
func WithID(id string) Option {
	return func(e *cloudevents.Event) error {
		e.SetID(id)
		return nil
	}
}

// WithAttributes sets additional attributes in the cloud event context
func WithAttributes(ceAttrs map[string]string) Option {
	return func(e *cloudevents.Event) error {
		for k, v := range ceAttrs {
			e.SetExtension(k, v)
		}
		return nil
	}
}

// NewFromVSphere returns a compliant CloudEvent for the given vSphere event
func NewFromVSphere(event types.BaseEvent, source string, options ...Option) (*cloudevents.Event, error) {
	eventInfo := GetDetails(event)
	ce := cloudevents.NewEvent(EventSpecVersion)

	// URI of the event producer, e.g. http(s)://vcenter.domain.ext/sdk
	ce.SetSource(source)

	// apply defaults
	ce.SetID(uuid.New().String())
	ce.SetTime(event.GetEvent().CreatedTime)

	ce.SetType(EventCanonicalType + "/" + eventInfo.Category)
	ce.SetSubject(eventInfo.Name)

	var err error
	err = ce.SetData(EventContentType, event)
	if err != nil {
		return nil, errors.Wrap(err, "set CloudEvent data")
	}

	// apply options
	for _, opt := range options {
		if err = opt(&ce); err != nil {
			return nil, errors.Wrap(err, "apply option")
		}
	}

	if err = ce.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation for CloudEvent failed")
	}

	return &ce, nil
}
