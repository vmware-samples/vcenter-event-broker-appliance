package function

import (
	"errors"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

// cloudEvent captures the event data
type cloudEvent struct {
	Data    types.Event
	Source  string
	Subject string
}

// pagerDutyData will be sent as a request to the pagerduty API
type pagerDutyData struct {
	RoutingKey  string `json:"routing_key"`
	EventAction string `json:"event_action"`
	Client      string `json:"client"`
	ClientURL   string `json:"client_url"`
	Payload     struct {
		Summary   string    `json:"summary"`
		Timestamp time.Time `json:"timestamp"`
		Source    string    `json:"source"`
		Severity  string    `json:"severity"`
		Component string    `json:"component"`
		Group     string    `json:"group"`
		Class     string    `json:"class"`
	} `json:"payload"`
}

// pdConfig is loaded from pdconfig json file
type pdConfig struct {
	RoutingKey  string `json:"routing_key"`
	EventAction string `json:"event_action"`
}

func validatePdConf(pdc pdConfig) error {
	if pdc.RoutingKey == "" {
		return errors.New("PagerDuty routing key cannot be empty")
	}

	if pdc.EventAction == "" {
		return errors.New("PagerDuty event action cannot be empty")
	}

	return nil
}

func isValidEvent(event cloudEvent) error {
	var msg string

	if event.Source == "" {
		msg = "invalid event: does not contain Source"
	}

	if event.Subject == "" {
		msg = "invalid event: does not contain Subject"
	}

	if event.Data.FullFormattedMessage == "" {
		msg = "invalid event: does not contain Data.FullFormattedMessage"
	}

	if (event.Data.CreatedTime == time.Time{}) {
		msg = "invalid event: does not contain Data.CreatedTime"
	}

	if event.Data.Vm.Name == "" {
		msg = "invalid event: does not contain Data.Vm.Name"
	}

	if event.Data.Host.Name == "" {
		msg = "invalid event: does not contain Data.Vm.Name"
	}

	if msg != "" {
		return errors.New(msg)
	}

	return nil
}
