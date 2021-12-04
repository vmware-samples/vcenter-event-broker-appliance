package tagging

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

const (
	tagActionAttach = "attach"
	tagActionDetach = "detach"
)

// Client runs a CloudEvents receiver and performs tagging operations.
type Client struct {
	soap       *govmomi.Client // SOAP
	rest       *rest.Client    // VAPI
	tagMgr     *tags.Manager
	cloudEvent client.Client // CloudEvents
	envConfig  EnvConfig
}

// EnvConfig collects the needed environment variables.
type EnvConfig struct {
	// Environment variable set by Knative.
	Port int `envconfig:"PORT" required:"true"`

	// vSphere settings.
	Insecure   bool   `envconfig:"VCENTER_INSECURE" default:"false"`
	VCAddress  string `envconfig:"VCENTER_URL" required:"true"`
	SecretPath string `envconfig:"VCENTER_SECRET_PATH" default:""`
	TagName    string `envconfig:"TAG_NAME" required:"true"`
	TagAction  string `envconfig:"TAG_ACTION" default:"attach"`

	DebugLogs bool `envconfig:"DEBUG" default:"true"`
}

// NewClient creates a client for vSphere, tagging, and CloudEvent methods.
func NewClient(ctx context.Context) (*Client, error) {
	var env EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, fmt.Errorf("process environment variables: %w", err)
	}

	sc, err := newSOAPClient(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("create vSphere SOAP client: %w", err)
	}

	rc, err := newRESTClient(ctx, sc.Client, env)
	if err != nil {
		return nil, fmt.Errorf("create vSphere REST client: %w", err)
	}
	tm := tags.NewManager(rc)

	ce, err := client.NewHTTP(cehttp.WithPort(env.Port))
	if err != nil {
		return nil, fmt.Errorf("create CloudEvents client: %w", err)
	}

	return &Client{
		soap:       sc,
		rest:       rc,
		cloudEvent: ce,
		tagMgr:     tm,
		envConfig:  env,
	}, nil
}

// Run starts the CloudEvents receiver.
func (c *Client) Run(ctx context.Context) error {
	logging.FromContext(ctx).Debugw("start CloudEvents receiver", zap.Int("port", c.envConfig.Port))
	return c.cloudEvent.StartReceiver(ctx, c.handler)
}

// Close releases resources used by the client.
func (c *Client) Close(ctx context.Context) {
	// Use fresh context to perform logout operations.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := c.soap.Logout(ctx); err != nil {
		logging.FromContext(ctx).Warnw("log out from vSphere SOAP API", zap.Error(err))
	}

	if err := c.rest.Logout(ctx); err != nil {
		logging.FromContext(ctx).Warnw("log out from vSphere REST API", zap.Error(err))
	}
}

func (c *Client) handler(ctx context.Context, event event.Event) protocol.Result {
	logging.FromContext(ctx).Debugw("received new event", zap.String("event", event.String()))

	// Retrieve the Managed Object Reference from the event.
	moRef, err := parseEventMoRef(event)
	if err != nil {
		return fmt.Errorf("parse CloudEvent: %w", err)
	}

	tag, err := c.tagMgr.GetTag(ctx, c.envConfig.TagName)
	if err != nil {
		return cehttp.NewResult(http.StatusBadRequest, "get tag %q: %w", c.envConfig.TagName, err)
	}

	switch c.envConfig.TagAction {
	case tagActionAttach:
		if err = c.tagMgr.AttachTag(ctx, tag.ID, moRef); err != nil {
			return fmt.Errorf("attach tagID %q to %q: %w", tag.ID, moRef.Value, err)
		}
		logging.FromContext(ctx).Debugw("attach tag to vm", zap.String("tag", tag.ID), zap.String("VM", moRef.Value))
	case tagActionDetach:
		if err = c.tagMgr.DetachTag(ctx, tag.ID, moRef); err != nil {
			return fmt.Errorf("detach tagID %q from %q: %w", tag.ID, moRef.Value, err)
		}
		logging.FromContext(ctx).Debugw("detach tag from vm", zap.String("tag", tag.ID), zap.String("VM", moRef.Value))
	}

	return nil
}

func parseEventMoRef(event event.Event) (*types.ManagedObjectReference, error) {
	var data types.Event
	if err := event.DataAs(&data); err != nil {
		return nil, err
	}
	if data.Vm == nil || data.Vm.Vm.Value == "" {
		return nil, errors.New("empty managed object reference")
	}

	moRef := types.ManagedObjectReference{
		Type:  data.Vm.Vm.Type,
		Value: data.Vm.Vm.Value,
	}
	return &moRef, nil
}
