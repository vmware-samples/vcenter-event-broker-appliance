package function

import (
	"io/ioutil"
	"testing"
)

const passMark = "\u2713"
const failMark = "\u2717"

// TestLoadTomlCfg shows valid vcconfig.toml files can be loaded and processed.
func TestLoadTomlCfg(t *testing.T) {
	var tests = []struct {
		testDesc  string
		cfgPath   string
		expectErr bool
		want      *vcConfig
	}{
		{
			"Test that toml file loads correctly",
			"testdata/vcconfig.toml",
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
				struct {
					URN    string
					Action string
				}{
					"urn:vmomi:InventoryServiceTag:11f16f36-f5c4-4c29-b7d3-d9c7d12babe6:GLOBAL",
					"attach",
				},
			},
		},
		{
			"Test that toml file loads, even with more info than needed, and defaults are set",
			"testdata/vcconfig2.toml",
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
				struct {
					URN    string
					Action string
				}{
					"urn:vmomi:InventoryServiceTag:11f16f36-f5c4-4c29-b7d3-d9c7d12babe6:GLOBAL",
					"detach",
				},
			},
		},
		{
			"Test that misconfigured toml file ends in error",
			"testdata/vcconfigErr1.toml",
			true,
			nil,
		},
		{
			"Test that vcconfig.toml missing essential information results in error.",
			"testdata/vcconfigErr2.toml",
			true,
			nil,
		},
		{
			"Test that missing toml file results in error",
			"testdata/missing.toml",
			true,
			nil,
		},
	}

	for _, tc := range tests {
		t.Logf("=========== %v ===========", tc.testDesc)
		cfg, err := loadTomlCfg(tc.cfgPath)
		if err != nil {
			if tc.expectErr {
				// An error is expected.
				t.Logf("got an error, as expected: %v. %v", err, passMark)
			} else {
				t.Log(tc.testDesc, failMark, err)
				t.Fail()
			}
		} else {
			if *cfg == *tc.want {
				t.Logf("got expected: %v. %v", tc.want, passMark)
			} else {
				t.Logf("expected: %v, got: %v. %v", tc.want, cfg, failMark)
				t.Fail()
			}
		}

	}
}

// TestParseEventMoRef ensures that managed object reference value and type are
// obtained by the event json that meets Cloud Event specifications.
func TestParseEventMoRef(t *testing.T) {
	type vm struct {
		vmType string
		Value  string
	}

	var tests = []struct {
		testDesc  string
		jsonPath  string
		expectErr bool
		want      *vm
	}{
		{
			"Test that event is readable",
			"testdata/event.json",
			false,
			&vm{vmType: "VirtualMachine", Value: "vm-10000"},
		},
		{
			"Event should be readable, even with minimal information",
			"testdata/event2.json",
			false,
			&vm{vmType: "VirtualMachine", Value: "vm-2"},
		},
		{
			"Event should return error if VM type and value are null",
			"testdata/eventErr1.json",
			true,
			nil,
		},
		{
			"Event should return error if VM info is null",
			"testdata/eventErr2.json",
			true,
			nil,
		},
		{
			"Event should return error if VM parent is null",
			"testdata/eventErr3.json",
			true,
			nil,
		},
		{
			"Event should return error if data is null",
			"testdata/eventErr4.json",
			true,
			nil,
		},
	}

	for _, tc := range tests {
		t.Logf("=========== %v ===========", tc.testDesc)
		body, err := ioutil.ReadFile(tc.jsonPath)
		if err != nil {
			t.Fatal("Test failing due to improper test setup.", failMark, err)
		}

		moRef, err := parseEventMoRef(body)
		if err != nil {
			if tc.expectErr {
				// An error is expected.
				t.Logf("got an error, as expected: %v. %v", err, passMark)
			} else {
				t.Log(tc.testDesc, failMark, err)
				t.Fail()
			}
		}

		if err == nil {
			if moRef.Type == tc.want.vmType {
				t.Logf("got expected: '%s'. %v", moRef.Type, passMark)
			} else {
				t.Logf("expected: '%s', got: '%s'. %v", tc.want.vmType, moRef.Type, passMark)
				t.Fail()
			}

			if moRef.Value == tc.want.Value {
				t.Logf("got expected: '%s'. %v", moRef.Value, passMark)
			} else {
				t.Fatalf("expected: '%s', got: '%s'. %v", tc.want.Value, moRef.Value, passMark)
			}
		}
	}
}
