package function

import (
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

// pdConfig is loaded from toml vcconfig file
type pdConfig struct {
	PagerDuty struct {
		RoutingKey  string `json:"routingKey"`
		EventAction string `json:"eventAction"`
	}
}

// cloudEvent is a subsection of a Cloud Event.
type cloudEvent struct {
	Data    types.Event `json:"data"`
	Source  string      `json:"source"`
	Subject string      `json:"subject"`
}

type pdPayload struct {
	RoutingKey  string `json:"routing_key"`
	EventAction string `json:"event_action"`
	Client      string `json:"client"`
	ClientURL   string `json:"client_url"`
	Payload     struct {
		Summary       string    `json:"summary"`
		Timestamp     time.Time `json:"timestamp"`
		Source        string    `json:"source"`
		Severity      string    `json:"severity"`
		Component     string    `json:"component"`
		Group         string    `json:"group"`
		Class         string    `json:"class"`
		CustomDetails struct {
			User            string                              `json:"user"`
			VM              *types.VmEventArgument              `json:"VM"`
			Host            *types.HostEventArgument            `json:"Host"`
			Datacenter      *types.DatacenterEventArgument      `json:"Datacenter"`
			ComputeResource *types.ComputeResourceEventArgument `json:"ComputeResource"`
		}
	} `json:"payload"`
}
