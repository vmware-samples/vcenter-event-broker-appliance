package processor

import (
	"log"
	"os"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

func Test_knativeProcessor_Process(t *testing.T) {
	var zero int = 0
	//var testLock sync.RWMutex
	var testLog *log.Logger = log.New(os.Stdout, color.Purple("[Knative_test] "), log.LstdFlags)
	type args struct {
		ce cloudevents.Event
	}
	tests := []struct {
		name       string
		kProcessor *KnativeProcessor
		args       args
		wantErr    bool
	}{
		{
			name: "Sending Sucessfull Cloud event to Knative Broker",
			kProcessor: &KnativeProcessor{
				address: "http://127.0.0.1:8080",
				verbose: false,
				client:  simpleBinaryClient("http://127.0.0.1:8080"),
				log:     testLog,

				stats: metrics.EventStats{EventsTotal: &zero, EventsErr: &zero},
			},
			args: args{ce: func() cloudevents.Event {
				e := cloudevents.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: time.Now()},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]string{
					"sq":  "42",
					"msg": "hello",
				})
				e.SetID("AABBCCDDEE")
				return e
			}()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.kProcessor.Process(tt.args.ce); (err != nil) != tt.wantErr {
				t.Errorf("knativeProcessor.Process() error = %v, wantErr %v", err, tt.wantErr)
			}

			if *tt.kProcessor.stats.EventsTotal != 1 {
				t.Errorf("knativeProcessor.Process() stats.EventsTotal = 1, want %v", *tt.kProcessor.stats.EventsTotal)
			}
		})
	}
}

func simpleBinaryClient(target string) client.Client {
	p, err := cehttp.New(cehttp.WithTarget(target))
	if err != nil {
		log.Printf("failed to create protocol, %v", err)
		return nil
	}

	c, err := client.New(p, client.WithForceBinary())
	if err != nil {
		return nil
	}
	return c
}
