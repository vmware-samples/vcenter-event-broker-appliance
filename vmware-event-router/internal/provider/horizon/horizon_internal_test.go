// +build unit

package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/jpillora/backoff"
	"go.uber.org/zap/zaptest"
	"gotest.tools/assert"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	fakeServer = "https://api.fake.horizon.com"
	testEvents = "./testdata/audit_events.golden"
)

func TestNewEventStream(t *testing.T) {
	type args struct {
		ctx  context.Context
		cfg  *config.ProviderConfigHorizon
		ms   metrics.Receiver
		log  logger.Logger
		opts []Option
	}
	tests := []struct {
		name      string
		args      args
		want      *EventStream
		errString string // err string assertion
	}{
		{
			name: "no credentials provided",
			args: args{
				ctx: context.TODO(),
				cfg: &config.ProviderConfigHorizon{
					Address: "https://myserver.horizon.com",
				},
				log: zaptest.NewLogger(t).Sugar(),
			},
			want:      nil,
			errString: fmt.Sprintf("invalid %s credentials:", config.ActiveDirectory),
		},
		{
			name: "no config provided",
			args: args{
				ctx: context.TODO(),
				cfg: nil,
				log: zaptest.NewLogger(t).Sugar(),
			},
			want:      nil,
			errString: "configuration must be provided",
		},
		{
			name: "invalid Horizon address provided",
			args: args{
				ctx: context.TODO(),
				cfg: &config.ProviderConfigHorizon{
					Address: "myserver//",
					Auth: &config.AuthMethod{
						Type: config.BasicAuth,
						BasicAuth: &config.BasicAuthMethod{
							Username: "user",
							Password: "pass",
						},
					},
				},
				log: zaptest.NewLogger(t).Sugar(),
			},
			want:      nil,
			errString: "invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEventStream(tt.args.ctx, tt.args.cfg, tt.args.ms, tt.args.log, tt.args.opts...)

			assert.ErrorContains(t, err, tt.errString)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func Test_removeDuplicates(t *testing.T) {
	t.Run("dup is nil", func(t *testing.T) {
		ev := createFakeEvents(10)
		got := removeDuplicates(ev, nil)
		assert.DeepEqual(t, got, ev)
	})

	t.Run("empty events", func(t *testing.T) {
		got := removeDuplicates([]AuditEventSummary{}, &AuditEventSummary{})
		assert.DeepEqual(t, got, []AuditEventSummary{})
	})

	t.Run("one duplicate event", func(t *testing.T) {
		ev := createFakeEvents(3)
		got := removeDuplicates(ev, &AuditEventSummary{ID: 10})
		assert.DeepEqual(t, got, ev[1:])
	})
}

// createFakeEvents creates returns a []AuditEventSummary where the ID of each
// element is set to the sum 10 plus the current counter
func createFakeEvents(count int) []AuditEventSummary {
	events := make([]AuditEventSummary, count)
	for i := 0; i < count; i++ {
		events[i] = AuditEventSummary{ID: int64(10 + i)}
	}

	return events
}

func TestEventStreamMock_Stream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	log := zaptest.NewLogger(t)

	f, err := os.Open(testEvents)
	assert.NilError(t, err, "open golden file: %s", testEvents)

	b, err := io.ReadAll(f)
	assert.NilError(t, err, "read golden file: %s", testEvents)

	var events []AuditEventSummary
	err = json.Unmarshal(b, &events)
	assert.NilError(t, err, "unmarshal golden file events")

	fc := fakeClient{
		events: events,
		log:    log.Sugar(),
	}

	stream := EventStream{
		client:       &fc,
		clock:        clock.New(),
		pollInterval: time.Millisecond * 10,
		Logger:       log.Sugar(),
		backoffConfig: &backoff.Backoff{
			Factor: 1,
			Jitter: false,
			Min:    0,
			Max:    time.Millisecond * 500,
		},
		stats: metrics.EventStats{
			Provider:    string(config.ProviderHorizon),
			Type:        config.EventProvider,
			Address:     fakeServer,
			Started:     time.Now().UTC(),
			EventsTotal: new(int),
			EventsErr:   new(int),
			EventsSec:   new(float64),
		},
	}

	fp := &fakeProcessor{
		t:      t,
		log:    log.Sugar(),
		expect: len(events), // there should be no duplicates
	}

	err = stream.Stream(ctx, fp)
	assert.ErrorContains(t, err, "context deadline exceeded")
	assert.Equal(t, fp.got, fp.expect)
}

type fakeClient struct {
	invocations int
	events      []AuditEventSummary
	log         logger.Logger
}

// GetEvents initially returns all events including up to second last events. On
// second invocation returns second last event. On third invocation returns
// second last and last event. Further invocations will only return last event.
func (f *fakeClient) GetEvents(_ context.Context, _ Timestamp) ([]AuditEventSummary, error) {
	f.invocations++
	f.log.Debugf("GetEvents invocations: %d", f.invocations)

	// preserve existing events slice
	var newEvents = make([]AuditEventSummary, len(f.events))
	copy(newEvents, f.events)

	// Horizon API returns events ordered from newest to oldest
	// note: concurrent events (by time) are not ordered by id (see golden file for example)
	switch f.invocations {
	case 1:
		// all up to (excluding) newest
		return newEvents[1:], nil
	case 2:
		// only second newest event
		return newEvents[1:2], nil
	case 3:
		// two newest events
		return newEvents[0:2], nil
	default:
		// only newest event
		return newEvents[0:1], nil
	}
}

func (f *fakeClient) Remote() string {
	return fakeServer
}

type fakeProcessor struct {
	t      *testing.T
	log    logger.Logger
	got    int
	expect int
}

func (f *fakeProcessor) Process(_ context.Context, ce ce.Event) error {
	f.log.Debugf("received new event: %s", ce.String())

	err := ce.Validate()
	assert.NilError(f.t, err)

	f.got++
	f.log.Debugf("processed events invocations: %d", f.got)
	return nil
}

func (f *fakeProcessor) PushMetrics(_ context.Context, _ metrics.Receiver) {}

func (f *fakeProcessor) Shutdown(_ context.Context) error {
	return nil
}

func Test_convertEventType(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "REST_AUTH_REFRESH_TOKEN_SUCCESS",
			args: args{t: "REST_AUTH_REFRESH_TOKEN_SUCCESS"},
			want: "com.vmware.event.router/horizon.rest_auth_refresh_token_success.v0",
		},
		{
			name: "VLSI_USERLOGGEDIN",
			args: args{t: "VLSI_USERLOGGEDIN"},
			want: "com.vmware.event.router/horizon.vlsi_userloggedin.v0",
		},
		{
			name: "REST_AUTH_LOGIN_SUCCESS",
			args: args{t: "REST_AUTH_LOGIN_SUCCESS"},
			want: "com.vmware.event.router/horizon.rest_auth_login_success.v0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEventType(tt.args.t)
			assert.Equal(t, got, tt.want)
		})
	}
}
