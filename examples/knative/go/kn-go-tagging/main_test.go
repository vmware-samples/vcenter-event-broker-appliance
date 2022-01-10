package main

import (
	"testing"

	"gotest.tools/assert"

	"github.com/vmware-samples/vcenter-event-broker-appliance/examples/knative/go/kn-go-tagging/tagging"
)

func TestValidateEnvConfig(t *testing.T) {
	tests := []struct {
		config  tagging.EnvConfig
		wantErr bool
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
		assert.Equal(t, tc.wantErr, err != nil)
	}
}
