package main

import (
	"testing"

	"github.com/vmware-samples/vcenter-event-broker-appliance/examples/knative/go/kn-go-tagging/tagging"
)

func TestValidateEnvConfig(t *testing.T) {
	tests := []struct {
		config tagging.EnvConfig
		isErr  bool
	}{
		{
			tagging.EnvConfig{TagAction: "attach"},
			false,
		},
		{
			tagging.EnvConfig{TagAction: "detach"},
			false,
		},
		{
			tagging.EnvConfig{TagAction: "dettach"},
			true,
		},
	}

	for _, tc := range tests {
		err := validateEnvConfig(tc.config)
		if tc.isErr && err == nil {
			t.Error("expected error but got none")
		}
		if !tc.isErr && err != nil {
			t.Errorf("did not expect error but got %v", err)
		}
	}
}
