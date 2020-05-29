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
	eventCanonicalType = "com.vmware.event.router"
	eventSpecVersion   = cloudevents.VersionV1
	eventContentType   = cloudevents.ApplicationJSON
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
func getDetails(event types.BaseEvent) (VCenterEventInfo, error) {
	eventInfo := VCenterEventInfo{}

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
	return eventInfo, nil
}

// NewCloudEvent returns a compliant CloudEvent
// TODO: make agnostic to just vCenter event types
func NewCloudEvent(event types.BaseEvent, source string) (*cloudevents.Event, error) {
	eventInfo, err := getDetails(event)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve event information")
	}

	ce := cloudevents.NewEvent(eventSpecVersion)

	// set ID of the event; must be non-empty and unique within the scope of the producer.
	ce.SetID(uuid.New().String())

	// set source - URI of the event producer, e.g. http(s)://vcenter.domain.ext/sdk.
	ce.SetSource(source)

	// set type - canonicalType + vcenter event category (event, eventex, extendedevent).
	ce.SetType(eventCanonicalType + "/" + eventInfo.Category)

	// set subject - vcenter event name used for topic subscriptions
	ce.SetSubject(eventInfo.Name)

	// set time - Timestamp set by this event router when this message was created.
	ce.SetTime(time.Now().UTC())

	// set data - Event payload as received from processor (includes event
	// creation timestamp, e.g. as set by vcenter).
	err = ce.SetData(eventContentType, event)
	if err != nil {
		return nil, errors.Wrap(err, "could not create CloudEvent")
	}

	if err = ce.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation for CloudEvent failed")
	}

	return &ce, nil
}
