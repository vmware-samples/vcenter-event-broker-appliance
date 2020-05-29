package metrics

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	// expvar map name for exposing the event router stats
	mapName = "vmware.event.router.stats"
	// PushInterval defines the default interval event streams and processors
	// push their metrics to the server
	PushInterval = time.Second * 1
)

// EventStats are provided and continously updated by event streams and
// processors
type EventStats struct {
	Provider     string         `json:"-"`             // ignored in JSON because provider is implicit via mapName[Provider]
	ProviderType string         `json:"provider_type"` // stream or processor
	Name         string         `json:"name"`
	Started      time.Time      `json:"started"`
	EventsTotal  *int           `json:"events_total,omitempty"`   // only used by event streams, total events received
	EventsErr    *int           `json:"events_err,omitempty"`     // only used by event streams, events received which lead to error
	EventsSec    *float64       `json:"events_per_sec,omitempty"` // only used by event streams
	Invocations  map[string]int `json:"invocations,omitempty"`    // event.Category to invocations - only used by event processors
}

func (s *EventStats) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		// will be printed to http stats endpoint
		return err.Error()
	}
	return string(b)
}

// load captures the 1/5/15 load interval of a GNU/Linux system
type load struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// function that will be called by expvar to export the information from the
// structure every time the endpoint is reached
func allLoadAvg() interface{} {
	return load{
		Load1:  loadAvg(0),
		Load5:  loadAvg(1),
		Load15: loadAvg(2),
	}
}

// helper function to retrieve the load average in GNU/Linux systems
func loadAvg(position int) float64 {
	// intentionally ignoring errors to make this work under non GNU/Linux
	// systems (testing)
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		//
		return 0
	}
	values := strings.Fields(string(data))
	load, err := strconv.ParseFloat(values[position], 64)
	if err != nil {
		return 0
	}
	return load
}
