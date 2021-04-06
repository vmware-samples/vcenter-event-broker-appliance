package function

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		desc      string
		path      string
		expectErr bool
		want      pdConfig
	}{
		{
			"test that valid pdconfig is parsed correctly",
			"testdata/valid-pdconfig.toml",
			false,
			pdConfig{
				PagerDuty: struct {
					RoutingKey  string `json:"routingKey"`
					EventAction string `json:"eventAction"`
				}{
					"amadeuptestroutingkey",
					"trigger",
				},
			},
		},
		{
			"test that missing routing key property will return error",
			"testdata/invalid-pdconfig-1.toml",
			true,
			pdConfig{},
		},
		{
			"test that missing trigger property will return error",
			"testdata/invalid-pdconfig-2.toml",
			true,
			pdConfig{},
		},
		{
			"test that incorrect toml syntax returns error",
			"testdata/invalid-pdconfig-3.toml",
			true,
			pdConfig{},
		},
	}

	for _, tc := range tests {
		got, err := loadConfig(tc.path)

		if tc.expectErr && err == nil {
			t.Fatalf("%s: want error but got none", tc.desc)
		}

		if !tc.expectErr {
			if err != nil {
				t.Fatalf("%s: load config: %v", tc.desc, err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
			}
		}
	}
}

func TestParseEvent(t *testing.T) {
	mockPDC := pdConfig{
		PagerDuty: struct {
			RoutingKey  string `json:"routingKey"`
			EventAction string `json:"eventAction"`
		}{
			"testroutingkey",
			"testtrigger",
		},
	}

	mockTime, err := time.Parse(time.RFC3339, "2020-09-15T21:03:03.183048Z")
	if err != nil {
		t.Fatalf("Setting mock time: %v", err)
	}

	tests := []struct {
		desc      string
		jsonPath  string
		expectErr bool
		want      pdPayload
	}{
		{
			"test that event missing data.Host and source properties return an error",
			"testdata/invalid-event-1.json",
			true,
			pdPayload{},
		},
		{
			"test that valid event is parsed correctly",
			"testdata/valid-event.json",
			false,
			pdPayload{
				RoutingKey:  "testroutingkey",
				EventAction: "testtrigger",
				Client:      "VMware Event Broker Appliance",
				ClientURL:   "https://veba.yourdomain.com/sdk",
				Payload: struct {
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
				}{
					Summary:   "Reconfigured core-A on esx-01a.yourdomain.com in RegionA01.  \n \nModified:  \n \nconfig.memoryHotAddEnabled: true -> false; \n\n Added:  \n \n Deleted:  \n \n",
					Timestamp: mockTime,
					Source:    "https://veba.yourdomain.com/sdk",
					Severity:  "info",
					Component: "core-A",
					Group:     "esx-01a.yourdomain.com",
					Class:     "VmReconfiguredEvent",
					CustomDetails: struct {
						User            string                              `json:"user"`
						VM              *types.VmEventArgument              `json:"VM"`
						Host            *types.HostEventArgument            `json:"Host"`
						Datacenter      *types.DatacenterEventArgument      `json:"Datacenter"`
						ComputeResource *types.ComputeResourceEventArgument `json:"ComputeResource"`
					}{
						User: "VSPHERE.LOCAL\\Administrator",
						Datacenter: &types.DatacenterEventArgument{
							EntityEventArgument: types.EntityEventArgument{
								Name: "RegionA01",
							},
							Datacenter: types.ManagedObjectReference{
								Type:  "Datacenter",
								Value: "datacenter-1001",
							},
						},
						ComputeResource: &types.ComputeResourceEventArgument{
							EntityEventArgument: types.EntityEventArgument{
								Name: "RegionA01-COMP01",
							},
							ComputeResource: types.ManagedObjectReference{
								Type:  "ClusterComputeResource",
								Value: "domain-c1006",
							},
						},
						Host: &types.HostEventArgument{
							EntityEventArgument: types.EntityEventArgument{
								Name: "esx-01a.yourdomain.com",
							},
							Host: types.ManagedObjectReference{
								Type:  "HostSystem",
								Value: "host-1009",
							},
						},
						VM: &types.VmEventArgument{
							EntityEventArgument: types.EntityEventArgument{
								Name: "core-A",
							},
							Vm: types.ManagedObjectReference{
								Type:  "VirtualMachine",
								Value: "vm-1047",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		body, err := ioutil.ReadFile(tc.jsonPath)
		if err != nil {
			t.Fatalf("%s: reading test data: %v", tc.desc, err)
		}

		got, err := parseEvent(body, mockPDC)

		if tc.expectErr && err == nil {
			t.Fatalf("%s: want error bot got none", tc.desc)
		}

		if !tc.expectErr {
			if err != nil {
				t.Fatalf("%s: parse cloud event: %v", tc.desc, err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s: got: %v, want: %v", tc.desc, got, tc.want)
			}
		}
	}
}
