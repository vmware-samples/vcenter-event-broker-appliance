package events

import (
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	eventCanonicalType = "com.vmware.event.router"
	eventSpecVersion   = "1.0" // CloudEvents spec version used
	eventContentType   = "application/json"
)

// CloudEvent is the JSON object sent to subscribed functions. We follow
// CloudEvents v1.0 spec as defined in
// https://github.com/cloudevents/sdk-go/blob/6c55828dbb6915e1594e5ace8bd8a19980731867/pkg/cloudevents/eventcontext_v1.go#L22
type CloudEvent struct {
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Source - URI of the event producer, e.g. http(s)://vcenter.domain.ext/sdk.
	Source string `json:"source"`
	// SpecVersion - The version of the CloudEvents specification the event router.
	SpecVersion string `json:"specversion"`
	// Type - canonicalType + vcenter event category (event, eventex, extendedevent).
	Type string `json:"type"`
	// Subject - vcenter event name used for topic subscriptions
	Subject string `json:"subject"`
	// Time - Timestamp set by this event router when this message was created.
	Time time.Time `json:"time"`
	// Data - Event payload as received from vcenter (includes event creation timestamp set by vcenter).
	Data types.BaseEvent `json:"data"`
	// DataContentType - A MIME (RFC2046) string describing the media type of `data`.
	DataContentType string `json:"datacontenttype"`
}

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
	eventInfo := VCenterEventInfo{}

	switch e := event.(type) {
	case *types.EventEx:
		eventInfo.Category = "eventex"
		eventInfo.Name = e.EventTypeId
	case *types.ExtendedEvent:
		eventInfo.Category = "extendedevent"
		eventInfo.Name = e.EventTypeId
	default:
		eType := reflect.TypeOf(event).Elem().Name()
		eventInfo.Category = "event"
		eventInfo.Name = eType
	}
	return eventInfo
}

// NewCloudEvent returns a compliant CloudEvent
func NewCloudEvent(event types.BaseEvent, eventInfo VCenterEventInfo, source string) CloudEvent {
	return CloudEvent{
		ID:              uuid.New().String(),
		Source:          source,
		SpecVersion:     eventSpecVersion,
		Type:            eventCanonicalType + "/" + eventInfo.Category,
		Subject:         eventInfo.Name,
		Time:            time.Now().UTC(),
		Data:            event,
		DataContentType: eventContentType,
	}
}
