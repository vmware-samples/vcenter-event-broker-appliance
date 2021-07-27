// +build unit

package horizon_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"

	"go.uber.org/zap/zaptest"
	"gotest.tools/assert"
	"knative.dev/pkg/logging"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/horizon"
)

const (
	// ENV_CONFIG_NAME contains the JSON-encoded horizonAPIConfig string
	EnvConfigName = "HORIZON_API_CONFIG"
)

// configuration struct passed via JSON-encoded string
type horizonAPIConfig struct {
	Address  string `json:"address,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
}

func TestEventStream_Stream(t *testing.T) {
	cfgJSON := os.Getenv(EnvConfigName)

	if cfgJSON == "" {
		t.Skipf("environment variable %q not set, skipping integration test", EnvConfigName)
	}

	var envCfg horizonAPIConfig
	err := json.Unmarshal([]byte(cfgJSON), &envCfg)
	assert.NilError(t, err)

	log := zaptest.NewLogger(t)
	ctx := logging.WithLogger(context.Background(), log.Sugar())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t.Run("receive events", func(t *testing.T) {
		cfg := config.ProviderConfigHorizon{
			Address:     envCfg.Address,
			InsecureSSL: envCfg.Insecure,
			Auth: &config.AuthMethod{
				Type: config.ActiveDirectory,
				ActiveDirectoryAuth: &config.ActiveDirectoryAuthMethod{
					Domain:   envCfg.Domain,
					Username: envCfg.Username,
					Password: envCfg.Password,
				},
			},
		}

		proc := fakeProcessor{
			t:      t,
			expect: 1, // events
			log:    log.Sugar(),
			doneCh: make(chan bool),
		}

		prov, err := horizon.NewEventStream(ctx, &cfg, &metricsStub{}, log.Sugar(), horizon.WithPollInterval(time.Second))
		assert.NilError(t, err)

		// stop test after timeout or expected number of events returned
		go func() {
			defer cancel()
			for {
				select {
				case <-proc.Done():
					return
				case <-time.After(time.Second * 5):
					t.Error("timed out")
					return
				}
			}
		}()

		err = prov.Stream(ctx, &proc)
		assert.ErrorContains(t, err, "context canceled")
	})
}

type metricsStub struct{}

func (m metricsStub) Receive(_ *metrics.EventStats) {}

type fakeProcessor struct {
	t      *testing.T
	log    logger.Logger
	doneCh chan bool
	got    int
	expect int
}

func (f *fakeProcessor) Process(_ context.Context, ce ce.Event) error {
	f.log.Debugf("received new event: %s", ce.String())

	err := ce.Validate()
	assert.NilError(f.t, err)

	f.got++
	if f.got == f.expect {
		close(f.doneCh)
	}

	return nil
}

func (f *fakeProcessor) Done() <-chan bool {
	return f.doneCh
}

func (f *fakeProcessor) PushMetrics(_ context.Context, _ metrics.Receiver) {}

func (f *fakeProcessor) Shutdown(_ context.Context) error {
	return nil
}
