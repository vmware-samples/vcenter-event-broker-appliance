//go:build unit

package aws

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	region   = "test-region"
	eventBus = "test-bus"
	ruleARN  = "test-rule"
)

type mockMetrics struct {
	sync.Mutex
	success int
	failed  int

	once     sync.Once
	received chan struct{} // signal Receive was called once
}

func (m *mockMetrics) Receive(stats *metrics.EventStats) {
	m.Lock()
	defer m.Unlock()

	if m.success != 0 || m.failed != 0 {
		m.once.Do(func() {
			//  close and signal if we received at least one real metric update
			close(m.received)
		})

		return
	}

	for _, v := range stats.Invocations {
		m.success += v.SuccessCount
		m.failed += v.FailureCount
	}
}

type mockClient struct {
	eventbridgeiface.EventBridgeAPI
	failPut  bool // returns generic failure on PutEvents
	failList bool // returns generic failure on ListRules
	pattern  string
	sent     int32 // track successful put calls
}

func (m *mockClient) PutEventsWithContext(_ aws.Context, input *eventbridge.PutEventsInput, _ ...request.Option) (*eventbridge.PutEventsOutput, error) {
	if m.failPut {
		return nil, fmt.Errorf("could not put event: %v", input)
	}

	atomic.AddInt32(&m.sent, 1)
	return &eventbridge.PutEventsOutput{}, nil
}

func (m *mockClient) ListRulesWithContext(_ aws.Context, input *eventbridge.ListRulesInput, _ ...request.Option) (*eventbridge.ListRulesOutput, error) {
	if m.failList {
		return nil, fmt.Errorf("could not list rules for input %v", input)
	}

	return &eventbridge.ListRulesOutput{
		NextToken: nil,
		Rules: []*eventbridge.Rule{
			{
				Arn:          aws.String(ruleARN),
				EventBusName: aws.String(eventBus),
				EventPattern: aws.String(m.pattern),
			},
		},
	}, nil
}

func TestEventBridgeProcessor_New(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		failList  bool   // fail list rule with error
		wantError string // expect send error
	}{
		{
			name:      "fails to create when rules cannot be listed",
			failList:  true,
			pattern:   `{"detail": {"subject": [{"exists":true}]}}`,
			wantError: "could not list",
		},
		{
			name:      "fails to create when pattern rule is empty",
			failList:  false,
			pattern:   "",
			wantError: "empty pattern",
		},
		{
			name:      "successfully creates processor",
			failList:  false,
			pattern:   `{"detail": {"subject": [{"exists":true}]}}`,
			wantError: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger := zaptest.NewLogger(t)

			cfg := config.ProcessorConfigEventBridge{
				Region:   region,
				EventBus: eventBus,
				RuleARN:  ruleARN,
			}

			ebClient := mockClient{
				pattern: tt.pattern,
			}

			ebClient.failList = tt.failList

			metricsClient := mockMetrics{
				received: make(chan struct{}),
			}

			_, err := NewEventBridgeProcessor(ctx, &cfg, &metricsClient, logger.Sugar(), WithClient(&ebClient))
			if tt.wantError != "" {
				assert.ErrorContains(t, err, tt.wantError)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestEventBridgeProcessor_Process(t *testing.T) {
	tests := []struct {
		name       string
		event      cloudevents.Event
		pattern    string
		wantSent   int
		wantFailed int
		wantError  string // expect send error
	}{
		{
			name:       "fails to send when PutEvents returns error",
			event:      newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:    `{"detail": {"subject": [{"exists":true}]}}`, // match any
			wantSent:   0,
			wantFailed: 1,
			wantError:  "could not put event",
		},
		{
			name:       "successfully sends event",
			event:      newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:    `{"detail": {"subject": [{"exists":true}]}}`, // match any
			wantSent:   1,
			wantFailed: 0,
			wantError:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger := zaptest.NewLogger(t)

			cfg := config.ProcessorConfigEventBridge{
				Region:   region,
				EventBus: eventBus,
				RuleARN:  ruleARN,
			}

			ebClient := mockClient{
				pattern: tt.pattern,
			}

			if tt.wantError != "" {
				ebClient.failPut = true
			}

			metricsClient := mockMetrics{
				received: make(chan struct{}),
			}

			proc, err := NewEventBridgeProcessor(ctx, &cfg, &metricsClient, logger.Sugar(), WithClient(&ebClient))
			assert.NilError(t, err)
			defer func() {
				err = proc.Shutdown(ctx)
				assert.NilError(t, err)
			}()

			err = proc.Process(ctx, tt.event)
			if tt.wantError != "" {
				assert.ErrorContains(t, err, tt.wantError)
			} else {
				assert.NilError(t, err)
			}

			<-metricsClient.received
			success := func() int {
				metricsClient.Lock()
				defer metricsClient.Unlock()
				return metricsClient.success
			}

			failed := func() int {
				metricsClient.Lock()
				defer metricsClient.Unlock()
				return metricsClient.failed
			}

			assert.Equal(t, tt.wantSent, success())
			assert.Equal(t, tt.wantFailed, failed())
		})
	}
}

func TestEventBridgeProcessor_Process_PatternMatch(t *testing.T) {
	tests := []struct {
		name     string
		event    cloudevents.Event
		pattern  string
		wantSent int32
	}{
		{
			name:     "pattern match on subject \"VmPoweredOnEvent\" or \"VmPoweredOffEvent\"",
			event:    newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:  `{"detail": {"subject": ["VmPoweredOnEvent","VmPoweredOffEvent"]}}`,
			wantSent: 1,
		},
		{
			name:     "pattern match on subject with prefix \"Vm\"",
			event:    newTestEvent(t, "VmReconfiguredEvent", newVMReconfiguredEvent()),
			pattern:  `{"detail": {"subject": [{"shellstyle":"Vm*"}]}}`,
			wantSent: 1,
		},
		{
			name:     "pattern match on VM name with prefix \"Linux\"",
			event:    newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:  `{"detail": {"data": {"Vm": {"Name": [{"shellstyle": "Linux*"}]}}}}`,
			wantSent: 1,
		},
		{
			name:     "pattern match on extended attribute eventclass anything-but \"eventex\" and \"extendedevent\"",
			event:    newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:  `{"detail": {"eventclass": [{"anything-but": ["extendedevent","eventex"]}]}}`,
			wantSent: 1,
		},
		{
			name:     "no pattern match on VM name with prefix \"Windows\"",
			event:    newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:  `{"detail": {"data": {"Vm": {"Name": [{"shellstyle": "Windows*"}]}}}}`,
			wantSent: 0,
		},
		{
			name:     "no pattern match on NULL Host value",
			event:    newTestEvent(t, "VmPoweredOnEvent", newVMPoweredOnEvent()),
			pattern:  `{"detail": {"data": {"Host": [{"exists": false}]}}}`,
			wantSent: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger := zaptest.NewLogger(t)

			cfg := config.ProcessorConfigEventBridge{
				Region:   region,
				EventBus: eventBus,
				RuleARN:  ruleARN,
			}

			ebClient := mockClient{
				pattern: tt.pattern,
			}

			proc, err := NewEventBridgeProcessor(ctx, &cfg, &mockMetrics{}, logger.Sugar(), WithClient(&ebClient))
			assert.NilError(t, err)
			defer func() {
				err = proc.Shutdown(ctx)
				assert.NilError(t, err)
			}()

			err = proc.Process(ctx, tt.event)
			assert.NilError(t, err)

			sent := atomic.LoadInt32(&ebClient.sent)
			assert.Equal(t, tt.wantSent, sent)
		})
	}
}

func newTestEvent(t *testing.T, subject string, data interface{}) cloudevents.Event {
	t.Helper()

	e := cloudevents.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("test.event.type.v0")
	e.SetSubject(subject)
	e.SetSource("test-source")
	e.SetExtension("eventclass", "event")

	err := e.SetData(cloudevents.ApplicationJSON, data)
	assert.NilError(t, err)

	return e
}

func newVMPoweredOnEvent() types.BaseEvent {
	return &types.VmPoweredOnEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Vm: &types.VmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: "Linux-1234",
					},
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1234",
					},
				},
			},
		},
	}
}

func newVMReconfiguredEvent() types.BaseEvent {
	return &types.VmReconfiguredEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Vm: &types.VmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: "Linux-1234",
					},
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1234",
					},
				},
			},
		},
	}
}
