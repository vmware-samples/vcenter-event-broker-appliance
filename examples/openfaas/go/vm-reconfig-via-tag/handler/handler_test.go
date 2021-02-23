package function

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// TestLoadConfig tests that the toml file loads correctly.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		desc      string
		cfgPath   string
		expectErr bool
		want      *vcConfig
	}{
		{
			"Toml file loads correctly, including default insecure",
			"testdata/valid-vcconfig-1.toml",
			false,
			&vcConfig{
				struct {
					Server   string
					User     string
					Password string
					Insecure bool
				}{
					"veba.local.corp",
					"admin@vsphere.local",
					"password1234",
					false,
				},
			},
		},
		{
			"Toml file loads, even with more info than needed",
			"testdata/valid-vcconfig-2.toml",
			false,
			&vcConfig{
				struct {
					Server   string
					User     string
					Password string
					Insecure bool
				}{
					"veba.local.corp",
					"admin@vsphere.local",
					"password1234",
					true,
				},
			},
		},
		{
			"Misconfigured toml file ends in error",
			"testdata/invalid-vcconfig-1.toml",
			true,
			nil,
		},
		{
			"Missing required information results in error.",
			"testdata/invalid-vcconfig-2.toml",
			true,
			nil,
		},
		{
			"Missing toml file results in error",
			"testdata/missing-vcconfig.toml",
			true,
			nil,
		},
	}

	for _, tc := range tests {
		got, err := loadConfig(tc.cfgPath)

		if tc.expectErr && err == nil {
			t.Fatalf("%s: want error, but got none", tc.desc)
		}

		if !tc.expectErr {
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.desc, err)
			}

			if *got != *tc.want {
				t.Fatalf("%s: got: %v, want: %v", tc.desc, got, tc.want)
			}
		}
	}
}

// TestEventVmMOR ensures the VM managed object reference is retrieved from cloud event.
func TestEventVmMOR(t *testing.T) {
	tests := []struct {
		desc      string
		jsonPath  string
		expectErr bool
		want      *types.ManagedObjectReference
	}{
		{
			"Cloud event data are readable",
			"testdata/valid-event-1.json",
			false,
			&types.ManagedObjectReference{Type: "VirtualMachine", Value: "vm-10000"},
		},
		{
			"Event should be readable, even with minimal information",
			"testdata/valid-event-2.json",
			false,
			&types.ManagedObjectReference{Type: "VirtualMachine", Value: "vm-2"},
		},
		{
			"Event should return error if VM type and value are null",
			"testdata/invalid-event-1.json",
			true,
			nil,
		},
		{
			"Event should return error if VM info is null",
			"testdata/invalid-event-2.json",
			true,
			nil,
		},
		{
			"Event should return error if VM parent is null",
			"testdata/invalid-event-3.json",
			true,
			nil,
		},
		{
			"Event should return error if data is null",
			"testdata/invalid-event-4.json",
			true,
			nil,
		},
	}

	for _, tc := range tests {
		body, err := ioutil.ReadFile(tc.jsonPath)
		if err != nil {
			t.Fatal("Test failing due to improper test setup.", err)
		}

		got, err := eventVmMOR(body)
		if tc.expectErr && err == nil {
			t.Fatalf("%s: want error, got none", tc.desc)
		}

		if !tc.expectErr {
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.desc, err)
			}

			if got.Type != tc.want.Type {
				t.Errorf("%s: got: %q, want: %q", tc.desc, got.Type, tc.want.Type)
			}

			if got.Value != tc.want.Value {
				t.Errorf("%s: got: %q, want: %q", tc.desc, got.Value, tc.want.Value)
			}
		}
	}
}

// TestCurrentConfigs tests that given a managed object VM, the correct slice of
// current configurations will be returned.
func TestCurrentConfigs(t *testing.T) {
	mockFalse := false
	mockTrue := true

	mockMoVM := mo.VirtualMachine{
		Config: &types.VirtualMachineConfigInfo{
			MemoryHotAddEnabled: &mockFalse,
			CpuHotAddEnabled:    &mockTrue,
			CpuHotRemoveEnabled: &mockFalse,
			Hardware: types.VirtualHardware{
				NumCPU:            2,
				NumCoresPerSocket: 1,
				MemoryMB:          1024,
			},
		},
	}

	tests := []struct {
		desc string
		want []vmConfig
	}{
		{
			"Given a managed object VM, return a slice containing current configurations",
			[]vmConfig{
				{
					name:   "numCPU",
					valInt: 2,
				},
				{
					name:   "memoryMB",
					valInt: 1024,
				},
				{
					name:   "numCoresPerSocket",
					valInt: 1,
				},
				{
					name:    "memoryHotAddEnabled",
					valBool: false,
				},
				{
					name:    "cpuHotRemoveEnabled",
					valBool: false,
				},
				{
					name:    "cpuHotRemoveEnabled",
					valBool: true,
				},
			},
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		got := currentConfigs(ctx, mockMoVM)

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
		}
	}
}

// TestBuildConfig tests that give a tag name (config value) and the config type,
// the vm config object will be returned.
func TestBuildConfig(t *testing.T) {
	tests := []struct {
		desc        string
		attachedTag tags.Tag
		configType  string
		want        vmConfig
	}{
		{
			"Build config for desired 3 CPU",
			tags.Tag{Name: "3"},
			"numCPU",
			vmConfig{
				name:   "numCPU",
				valInt: 3,
			},
		},
		{
			"Build config for desired 8gb RAM",
			tags.Tag{Name: "8192"},
			"memoryMB",
			vmConfig{
				name:   "memoryMB",
				valInt: 8192,
			},
		},
		{
			"Build config for desired 2 cores per socket",
			tags.Tag{Name: "2"},
			"numCoresPerSocket",
			vmConfig{
				name:   "numCoresPerSocket",
				valInt: 2,
			},
		},
		{
			"Build config for desired memoryMB hot add enabled true",
			tags.Tag{Name: "true"},
			"memoryHotAddEnabled",
			vmConfig{
				name:    "memoryHotAddEnabled",
				valBool: true,
			},
		},
		{
			"Build config for desired CPU hot remove enabled true",
			tags.Tag{Name: "true"},
			"cpuHotRemoveEnabled",
			vmConfig{
				name:    "cpuHotRemoveEnabled",
				valBool: true,
			},
		},
		{
			"Build config for desired CPU hot add enabled false",
			tags.Tag{Name: "false"},
			"cpuHotAddEnabled",
			vmConfig{
				name:    "cpuHotAddEnabled",
				valBool: false,
			},
		},
	}

	for _, tc := range tests {
		got, err := buildConfig(tc.attachedTag, tc.configType)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.desc, err)
		}

		if got != tc.want {
			t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
		}
	}
}

// TestFilterOutCurrentConfigs will test that unapplied configs are tags that are
// in the attached tags but not in the current configs list.
func TestFilterOutCurrentConfigs(t *testing.T) {
	tests := []struct {
		desc     string
		tagCfgs  []vmConfig
		currCfgs []vmConfig
		want     []vmConfig
	}{
		{
			"Determine unapplied configs given configs in tags and current configs",
			[]vmConfig{
				{
					name:   "numCPU",
					valInt: 8,
				},
				{
					name:   "memoryMB",
					valInt: 16384,
				},
				{
					name:    "memoryHotAddEnabled",
					valBool: false,
				},
				{
					name:    "cpuHotAddEnabled",
					valBool: true,
				},
			},
			[]vmConfig{
				{
					name:   "numCPU",
					valInt: 8,
				},
				{
					name:   "memoryMB",
					valInt: 8192,
				},
				{
					name:   "numCoresPerSocket",
					valInt: 4,
				},
				{
					name:    "memoryHotAddEnabled",
					valBool: false,
				},
				{
					name:    "cpuHotRemoveEnabled",
					valBool: false,
				},
				{
					name:    "cpuHotAddEnabled",
					valBool: false,
				},
			},
			[]vmConfig{
				{
					name:   "memoryMB",
					valInt: 16384,
				},
				{
					name:    "cpuHotAddEnabled",
					valBool: true,
				},
			},
		},
	}

	for _, tc := range tests {
		got := filterOutCurrentConfigs(tc.tagCfgs, tc.currCfgs)

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
		}
	}
}

// TestGenerateDesiredSpec will test that the correct VM configuration spec is
// being generated based on unapplied configs and spec change version.
func TestGenerateDesiredSpec(t *testing.T) {
	mockMoVM := mo.VirtualMachine{
		Config: &types.VirtualMachineConfigInfo{
			ChangeVersion: "2020-09-14T18:53:02.85686Z",
		},
	}
	mockFalse := false

	tests := []struct {
		desc string
		cfgs []vmConfig
		want types.VirtualMachineConfigSpec
	}{
		{
			"Generated desired specs contains configurations to apply",
			[]vmConfig{
				{
					name:   "memoryMB",
					valInt: 1024,
				},
				{
					name:   "numCoresPerSocket",
					valInt: 1,
				},
				{
					name:    "cpuHotRemoveEnabled",
					valBool: false,
				},
			},
			types.VirtualMachineConfigSpec{
				ChangeVersion:       "2020-09-14T18:53:02.85686Z",
				MemoryMB:            1024,
				NumCoresPerSocket:   1,
				NumCPUs:             0,
				CpuHotRemoveEnabled: &mockFalse,
				CpuHotAddEnabled:    nil,
				MemoryHotAddEnabled: nil,
			},
		},
	}

	for _, tc := range tests {
		got := generateDesiredSpec(tc.cfgs, mockMoVM)

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
		}
	}
}
