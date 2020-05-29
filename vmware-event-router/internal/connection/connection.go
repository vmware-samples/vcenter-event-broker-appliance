package connection

import (
	"encoding/json"
	"io"
)

// Configs is a list of configurations for stream processors, providers and
// internal services of the VMware Event Router (e.g. metrics)
type Configs []Config

// Config is used to configure stream processors, providers and internal
// services of the VMware Event Router (e.g. metrics)
type Config struct {
	Type     string            `json:"type,omitempty"`     // "stream", "processor"
	Provider string            `json:"provider,omitempty"` // "vmware_vcenter", "openfaas", "aws_event_bridge"
	Address  string            `json:"address,omitempty"`
	Auth     Authentication    `json:"auth,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// Authentication can hold generic authentication data for different stream
// providers and processors
type Authentication struct {
	Method string            `json:"method"`
	Secret map[string]string `json:"secret"`
}

// Parse parses a list of configurations
func Parse(cfg io.Reader) (Configs, error) {
	var cfgs Configs
	err := json.NewDecoder(cfg).Decode(&cfgs)
	if err != nil {
		return nil, err
	}
	return cfgs, nil
}
