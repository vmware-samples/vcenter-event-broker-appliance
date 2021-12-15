package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	preemption "github.com/embano1/vsphere-preemption"
	errs "github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
	"github.com/vmware/govmomi/vim25/types"
	filterpb "go.temporal.io/api/filter/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdk "go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

const (
	maxLastWorkflows   = 10 // retrieve up to workflows for cancellation
	wfExecutionTimeout = time.Hour * 24
)

type envConfig struct {
	// Temporal settings
	Address   string `envconfig:"TEMPORAL_URL" required:"true"`
	Namespace string `envconfig:"TEMPORAL_NAMESPACE" required:"true"`
	Queue     string `envconfig:"TEMPORAL_TASKQUEUE" required:"true"`

	// vsphere settings
	Tag       string `envconfig:"VSPHERE_PREEMPTION_TAG" required:"true"` // vsphere tag
	AlarmName string `envconfig:"VSPHERE_ALARM_NAME" required:"true"`     // vsphere alarm name

	// Knative settings (injected)
	Sink string `envconfig:"K_SINK" required:"false"` // via sinkbinding (optional)
	Port int    `envconfig:"PORT" required:"true"`

	Debug bool `envconfig:"DEBUG" default:"false"`
}

type client struct {
	tc sdk.Client

	address   string // temporal address
	namespace string // temporal namespace
	queue     string // temporal queue
	tag       string // identifies preemptible vms
	alarmName string // identifies alarm to trigger workflow
	sink      string // K_SINK binding injection
}

func newClient(_ context.Context, logger *zap.Logger) (*client, error) {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}

	if env.Sink == "" {
		logger.Info("K_SINK variable not set")
	}

	tc, err := sdk.NewClient(sdk.Options{
		HostPort:  env.Address,
		Namespace: env.Namespace,
		Logger:    preemption.NewZapAdapter(logger),
	})
	if err != nil {
		return nil, err
	}

	c := client{
		tc:        tc,
		address:   env.Address,
		namespace: env.Namespace,
		queue:     env.Queue,
		tag:       env.Tag,
		alarmName: env.AlarmName,
		sink:      env.Sink,
	}
	return &c, nil
}

func (c *client) handler(ctx context.Context, event ce.Event) error {
	logger := logging.FromContext(ctx).With("eventID", event.ID())
	logger.Debugw("received event", "event", event.String())

	var ae types.AlarmStatusChangedEvent
	err := event.DataAs(&ae)
	if err != nil {
		logger.Warnw("get alarm event data", zap.Error(err))
		return ce.NewHTTPResult(http.StatusBadRequest, "event payload must be valid AlarmStatusChangedEvent")
	}

	if ae.Alarm.Name != c.alarmName {
		logger.Debugw("alarm event name does not match, skipping event", "incomingAlarmName", ae.Alarm.Name, "definedAlarmName", c.alarmName)
		return nil
	}

	changedFrom := strings.ToLower(ae.From)
	changedTo := strings.ToLower(ae.To)

	// check whether this was a valid AlarmStatusChangedEvent and not other
	// (inherited) AlarmEvent type
	if changedTo == "" || changedFrom == "" {
		logger.Warn("event is not of type AlarmStatusChangedEvent")
		return ce.NewHTTPResult(http.StatusBadRequest, "event payload must be valid AlarmStatusChangedEvent")
	}

	criticality := func() preemption.Criticality {
		switch changedTo {
		case "yellow":
			return preemption.CriticalityMedium
		case "red":
			return preemption.CriticalityHigh
		default:
			// 	treat any other (even unset) value as LOW
			return preemption.CriticalityLow
		}
	}

	// retrieve in-progress workflows to avoid multiple executions (best effort
	// due to concurrent function executions and lack of conditional workflow
	// execution)
	running, err := c.getRunningWorkflows(ctx)
	if err != nil {
		// just warn and continue
		logger.Warnw("get running workflows", zap.Error(err))
	}

	if isAlarmRaising(changedFrom, changedTo) {
		logger.Infow("alarm level is raising", "from", changedFrom, "to", changedTo)
		logger.Infow("triggering preemption")
		if err = c.triggerPreemption(ctx, event, criticality()); err != nil {
			logger.Errorw("trigger preemption", zap.Error(err))
			return ce.NewHTTPResult(http.StatusInternalServerError, "failed to run preemption")
		}
		return nil
	}

	logger.Infow("alarm level is decreasing", "from", changedFrom, "to", changedTo)
	logger.Infow("canceling any running workflows", "running", len(running))

	if err = c.cancelRunningWorkflows(ctx, running); err != nil {
		// just log and return successfully (no retry)
		logger.Warnw("cancel running workflows", zap.Error(err))
	}
	return nil
}

func isAlarmRaising(changedFrom string, changedTo string) bool {
	return (changedFrom != "red" && changedTo == "yellow") || changedTo == "red"
}

func (c *client) getRunningWorkflows(ctx context.Context) ([]*workflow.WorkflowExecutionInfo, error) {
	filter := workflowservice.ListOpenWorkflowExecutionsRequest_TypeFilter{
		TypeFilter: &filterpb.WorkflowTypeFilter{
			Name: preemption.WorkflowName,
		},
	}
	wfs, err := c.tc.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
		Namespace:       c.namespace,
		MaximumPageSize: maxLastWorkflows,
		Filters:         &filter,
	})
	if err != nil {
		return nil, err
	}

	return wfs.Executions, nil
}

func (c *client) triggerPreemption(ctx context.Context, e ce.Event, criticality preemption.Criticality) error {
	req := preemption.WorkflowRequest{
		Tag:         c.tag,
		Event:       e,
		Criticality: criticality,
		ReplyTo:     c.sink,
	}

	options := sdk.StartWorkflowOptions{
		ID:                       c.alarmName,
		TaskQueue:                c.queue,
		WorkflowExecutionTimeout: wfExecutionTimeout,
		// WorkflowIDReusePolicy:
		// enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE, // multiple
		// executions handled in workflow
		WorkflowExecutionErrorWhenAlreadyStarted: false,
	}

	// fire and forget
	logging.FromContext(ctx).Infow(
		"executing workflow",
		zap.String("workflow", preemption.WorkflowName),
		zap.String("eventID", e.ID()),
		zap.String("queue", c.queue),
		zap.String("tag", c.tag),
		zap.String("criticality", string(criticality)),
		zap.String("sink", c.sink),
		zap.String("alarmName", c.alarmName),
		zap.String("event", e.String()),
	)

	// alarm name is used as the workflow name and a new workflow is started
	// unless it is already running
	if _, err := c.tc.SignalWithStartWorkflow(ctx, c.alarmName, preemption.SignalChannel, req, options, preemption.WorkflowName); err != nil {
		return err
	}
	return nil
}

func (c *client) cancelRunningWorkflows(ctx context.Context, running []*workflow.WorkflowExecutionInfo) error {
	var cancelErrs error
	for _, wf := range running {
		id := wf.Execution.WorkflowId
		runID := wf.Execution.RunId
		logging.FromContext(ctx).Debugw("canceling workflow", zap.String("ID", id), zap.String("runID", runID))
		if err := c.tc.CancelWorkflow(ctx, id, runID); err != nil {
			cancelErrs = errs.Append(cancelErrs, err)
		}
	}

	return cancelErrs
}

func (c *client) close() {
	c.tc.Close()
}
