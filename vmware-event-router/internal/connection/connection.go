package connection

import (
	"encoding/json"
	"io"
)

type Configs []Config

type Config struct {
	Type     string            `json:"type,omitempty"`     // "stream", "processor"
	Provider string            `json:"provider,omitempty"` // "vmware_vcenter", "openfaas", "aws_event_bridge"
	Address  string            `json:"address,omitempty"`
	Auth     Authentication    `json:"auth,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

type Authentication struct {
	Method string            `json:"method"`
	Secret map[string]string `json:"secret"`
}

func Parse(cfg io.Reader) (Configs, error) {
	var cfgs Configs
	err := json.NewDecoder(cfg).Decode(&cfgs)
	if err != nil {
		return nil, err
	}
	return cfgs, nil
}
