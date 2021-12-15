package main

import (
	"context"
	"errors"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	preemption "github.com/embano1/vsphere-preemption"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/govmomi/vim25/types"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdk "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"
	"knative.dev/pkg/logging"
)

var any = mock.Anything

func Test_isAlarmRaising(t *testing.T) {
	type args struct {
		changedFrom string
		changedTo   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Alarm from grey to green", args: args{changedFrom: "grey", changedTo: "green"}, want: false},
		{name: "Alarm from grey to yellow", args: args{changedFrom: "grey", changedTo: "yellow"}, want: true},
		{name: "Alarm from red to green", args: args{changedFrom: "red", changedTo: "green"}, want: false},
		{name: "Alarm from green to yellow", args: args{changedFrom: "green", changedTo: "yellow"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlarmRaising(tt.args.changedFrom, tt.args.changedTo); got != tt.want {
				t.Errorf("isAlarmRaising() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_handler(t *testing.T) {
	t.Run("returns http 400 error if not AlarmStatusChanged event", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		incomingAlarm := "test-alarm"
		configuredAlarm := "test-alarm"

		// not AlarmStatusChangedEvent missing from/to
		alarmEvent := types.AlarmEvent{
			Alarm: types.AlarmEventArgument{
				EntityEventArgument: types.EntityEventArgument{
					Name: incomingAlarm,
				},
			},
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.Equal(t, err.(*http.Result).StatusCode, nethttp.StatusBadRequest)
		mc.AssertExpectations(t)
	})

	t.Run("skips workflow with AlarmStatusChanged not matching", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		incomingAlarm := "some-alarm"
		configuredAlarm := "test-alarm"

		alarmEvent := types.AlarmStatusChangedEvent{
			AlarmEvent: types.AlarmEvent{
				Alarm: types.AlarmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: incomingAlarm,
					},
				},
			},
			From: "green",
			To:   "red",
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.NilError(t, err)
		mc.AssertExpectations(t)
	})

	t.Run("triggers workflow with criticality HIGH from AlarmStatusChanged event raising from green to red", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		mc.On("ListOpenWorkflow", any, any).Return(&workflowservice.ListOpenWorkflowExecutionsResponse{}, nil)
		mc.On("SignalWithStartWorkflow", any, "test-alarm", any, any, any, any).Return(&mocks.WorkflowRun{}, nil)

		incomingAlarm := "test-alarm"
		configuredAlarm := "test-alarm"

		alarmEvent := types.AlarmStatusChangedEvent{
			AlarmEvent: types.AlarmEvent{
				Alarm: types.AlarmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: incomingAlarm,
					},
				},
			},
			From: "green",
			To:   "red",
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.NilError(t, err)

		gotCriticality := mc.Calls[1].Arguments.Get(3).(preemption.WorkflowRequest).Criticality
		wantCriticality := preemption.CriticalityHigh
		assert.Equal(t, wantCriticality, gotCriticality)

		mc.AssertExpectations(t)
	})

	t.Run("triggers workflow with criticality MEDIUM from AlarmStatusChanged event raising from green to yellow", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		mc.On("ListOpenWorkflow", any, any).Return(&workflowservice.ListOpenWorkflowExecutionsResponse{}, nil)
		mc.On("SignalWithStartWorkflow", any, "test-alarm", any, any, any, any).Return(&mocks.WorkflowRun{}, nil)

		incomingAlarm := "test-alarm"
		configuredAlarm := "test-alarm"

		alarmEvent := types.AlarmStatusChangedEvent{
			AlarmEvent: types.AlarmEvent{
				Alarm: types.AlarmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: incomingAlarm,
					},
				},
			},
			From: "green",
			To:   "yellow",
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.NilError(t, err)

		gotCriticality := mc.Calls[1].Arguments.Get(3).(preemption.WorkflowRequest).Criticality
		wantCriticality := preemption.CriticalityMedium
		assert.Equal(t, wantCriticality, gotCriticality)

		mc.AssertExpectations(t)
	})

	t.Run("cancels running workflow from AlarmStatusChanged event dropping from red to green", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		mc.On("ListOpenWorkflow", any, any).Return(&workflowservice.ListOpenWorkflowExecutionsResponse{
			Executions: []*workflow.WorkflowExecutionInfo{
				{Execution: &common.WorkflowExecution{
					WorkflowId: "wf-id-1",
					RunId:      "wf-id-1-run-1",
				}},
			},
		}, nil)

		mc.On("CancelWorkflow", any, "wf-id-1", "wf-id-1-run-1").Return(nil)

		incomingAlarm := "test-alarm"
		configuredAlarm := "test-alarm"

		alarmEvent := types.AlarmStatusChangedEvent{
			AlarmEvent: types.AlarmEvent{
				Alarm: types.AlarmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: incomingAlarm,
					},
				},
			},
			From: "red",
			To:   "green",
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.NilError(t, err)

		mc.AssertExpectations(t)
	})

	t.Run("fails to trigger workflow due to temporal error", func(t *testing.T) {
		t.Parallel()
		ctx := logging.WithLogger(context.Background(), zaptest.NewLogger(t).Sugar())
		mc := new(mocks.Client)

		mc.On("ListOpenWorkflow", any, any).Return(&workflowservice.ListOpenWorkflowExecutionsResponse{}, nil)
		mc.On("SignalWithStartWorkflow", any, "test-alarm", any, any, any, any).Return(nil, errors.New("internal error"))

		incomingAlarm := "test-alarm"
		configuredAlarm := "test-alarm"

		alarmEvent := types.AlarmStatusChangedEvent{
			AlarmEvent: types.AlarmEvent{
				Alarm: types.AlarmEventArgument{
					EntityEventArgument: types.EntityEventArgument{
						Name: incomingAlarm,
					},
				},
			},
			From: "green",
			To:   "red",
		}

		e := event.New()
		e.SetID("1")
		e.SetSource("https://vcenter.local/sdk")
		e.SetType("mock.vsphere.AlarmStatusChangedEvent.v0")
		e.SetTime(time.Now().UTC())
		err := e.SetData(event.ApplicationJSON, alarmEvent)
		assert.NilError(t, err)

		c := newTestClient(mc, configuredAlarm)
		err = c.handler(ctx, e)
		assert.Equal(t, err.(*http.Result).StatusCode, nethttp.StatusInternalServerError)
		mc.AssertExpectations(t)
	})
}

func newTestClient(tc sdk.Client, alarm string) client {
	return client{
		tc:        tc,
		address:   "https://temporal.local",
		namespace: "temporal-test",
		queue:     "temporal-test",
		tag:       "test-tag",
		alarmName: alarm,
		sink:      "",
	}
}
