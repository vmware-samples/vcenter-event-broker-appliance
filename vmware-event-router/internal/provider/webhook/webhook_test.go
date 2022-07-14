//go:build unit

package webhook_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	ce "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"gotest.tools/v3/assert"
	"knative.dev/pkg/logging"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/webhook"
)

func Test_WebhookServer(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	t.Run("fails to start with invalid webhook config", func(t *testing.T) {
		tests := []struct {
			name      string
			address   string
			path      string
			errString string
		}{
			{"root path is not allowed", "127.0.0.1:0", "/", webhook.ErrInvalidPath.Error()},
			{"invalid bind address", "abc:0", "", "invalid webhook config: invalid character detected"},
		}

		for _, tt := range tests {
			test := tt
			t.Run(test.name, func(t *testing.T) {
				cfg := config.ProviderConfigWebhook{
					BindAddress: test.address,
					Path:        test.path,
				}

				_, err := webhook.NewServer(context.TODO(), &cfg, metricsStub{}, logger.Sugar())
				assert.ErrorContains(t, err, test.errString)
			})
		}
	})

	t.Run("send event to server with custom endpoint", func(t *testing.T) {
		tests := []struct {
			name    string
			address string
			path    string
			wantErr bool
		}{
			{"path /my-webhook", "127.0.0.1:0", "/webhook", false},
			{"path /receiver", "127.0.0.1:0", "/receiver", false},
		}

		for _, tt := range tests {
			test := tt
			t.Run(test.name, func(t *testing.T) {
				cfg := config.ProviderConfigWebhook{
					BindAddress: "127.0.0.1:0",
					Path:        test.path,
				}

				ctx := logging.WithLogger(context.Background(), logger.Sugar())
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				srv, err := webhook.NewServer(ctx, &cfg, metricsStub{}, logger.Sugar())
				assert.NilError(t, err, "run server")

				var eg errgroup.Group

				// cloud event sender
				eg.Go(func() error {
					defer cancel()

					target := fmt.Sprintf("http://%s%s", srv.Address(), test.path)
					res := sendEvent(ctx, t, target, "")

					var httpResult *cehttp.Result
					ce.ResultAs(res, &httpResult)
					assert.Equal(t, httpResult.StatusCode, 200)
					assert.Equal(t, ce.IsUndelivered(res), false, "Failed to send: %v", res)
					assert.Equal(t, ce.IsNACK(res), false, "Failed to send: %v", res)

					return nil
				})

				err = srv.Stream(ctx, &fakeProcessor{logger.Sugar()})
				assert.NilError(t, err, "run server")

				err = eg.Wait()
				assert.NilError(t, err, "http client")
			})
		}
	})

	t.Run("send event to server with default endpoint", func(t *testing.T) {
		cfg := config.ProviderConfigWebhook{
			BindAddress: "127.0.0.1:0",
			Path:        "",
		}

		ctx := logging.WithLogger(context.Background(), logger.Sugar())
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		srv, err := webhook.NewServer(ctx, &cfg, metricsStub{}, logger.Sugar())
		assert.NilError(t, err, "run server")

		var eg errgroup.Group

		// cloud event sender
		eg.Go(func() error {
			defer cancel()

			target := fmt.Sprintf("http://%s/webhook", srv.Address())
			res := sendEvent(ctx, t, target, "")

			var httpResult *cehttp.Result
			ce.ResultAs(res, &httpResult)
			assert.Equal(t, httpResult.StatusCode, 200)
			assert.Equal(t, ce.IsUndelivered(res), false, "Failed to send: %v", res)
			assert.Equal(t, ce.IsNACK(res), false, "Failed to send: %v", res)

			return nil
		})

		err = srv.Stream(ctx, &fakeProcessor{logger.Sugar()})
		assert.NilError(t, err, "run server")

		err = eg.Wait()
		assert.NilError(t, err, "http client")
	})

	t.Run("send event to server with basic auth", func(t *testing.T) {
		cfg := config.ProviderConfigWebhook{
			BindAddress: "127.0.0.1:0",
			Path:        "",
			Auth: &config.AuthMethod{
				Type: config.BasicAuth,
				BasicAuth: &config.BasicAuthMethod{
					Username: "user",
					Password: "pass",
				},
			},
		}

		ctx := logging.WithLogger(context.Background(), logger.Sugar())
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		srv, err := webhook.NewServer(ctx, &cfg, metricsStub{}, logger.Sugar())
		assert.NilError(t, err, "run server")

		var eg errgroup.Group

		// cloud event sender
		eg.Go(func() error {
			defer cancel()

			target := fmt.Sprintf("http://%s/webhook", srv.Address())
			res := sendEvent(ctx, t, target, basicAuth("user", "pass"))

			var httpResult *cehttp.Result
			ce.ResultAs(res, &httpResult)
			assert.Equal(t, httpResult.StatusCode, 200)
			assert.Equal(t, ce.IsUndelivered(res), false, "Failed to send: %v", res)
			assert.Equal(t, ce.IsNACK(res), false, "Failed to send: %v", res)

			return nil
		})

		err = srv.Stream(ctx, &fakeProcessor{logger.Sugar()})
		assert.NilError(t, err, "run server")

		err = eg.Wait()
		assert.NilError(t, err, "http client")
	})

	t.Run("send event without credentials to server with basic auth", func(t *testing.T) {
		cfg := config.ProviderConfigWebhook{
			BindAddress: "127.0.0.1:0",
			Path:        "",
			Auth: &config.AuthMethod{
				Type: config.BasicAuth,
				BasicAuth: &config.BasicAuthMethod{
					Username: "user",
					Password: "pass",
				},
			},
		}

		ctx := logging.WithLogger(context.Background(), logger.Sugar())
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		srv, err := webhook.NewServer(ctx, &cfg, metricsStub{}, logger.Sugar())
		assert.NilError(t, err, "run server")

		var eg errgroup.Group

		// cloud event sender
		eg.Go(func() error {
			defer cancel()

			target := fmt.Sprintf("http://%s/webhook", srv.Address())
			res := sendEvent(ctx, t, target, "")

			var httpResult *cehttp.Result
			ce.ResultAs(res, &httpResult)
			assert.Equal(t, httpResult.StatusCode, 401)
			assert.Equal(t, ce.IsUndelivered(res), false, "Failed to send: %v", res)
			assert.Equal(t, ce.IsNACK(res), true, "Failed to send: %v", res)

			return nil
		})

		err = srv.Stream(ctx, &fakeProcessor{logger.Sugar()})
		assert.NilError(t, err, "run server")

		err = eg.Wait()
		assert.NilError(t, err, "http client")
	})
}

func sendEvent(ctx context.Context, t *testing.T, target string, creds string) error {
	ctx = ce.ContextWithTarget(ctx, target)

	var httpProtocol *cehttp.Protocol
	if creds != "" {
		p, ceErr := ce.NewHTTP(ce.WithHeader("Authorization", "Basic "+creds))
		assert.NilError(t, ceErr, "create cloud event transport")
		httpProtocol = p
	} else {
		p, ceErr := ce.NewHTTP()
		assert.NilError(t, ceErr, "create cloud event transport")
		httpProtocol = p
	}

	c, ceErr := ce.NewClient(httpProtocol, ce.WithTimeNow(), ce.WithUUIDs())
	assert.NilError(t, ceErr, "create cloud event transport")

	e := ce.NewEvent()
	e.SetType("com.ce.sample.sent")
	e.SetSource("https://https://github.com/vmware-samples/vcenter-event-broker-appliance/sender")
	err := e.SetData(ce.ApplicationJSON, map[string]interface{}{
		"id":      1,
		"message": "Hello, World!",
	})
	assert.NilError(t, err, "set cloud event data")

	return c.Send(ctx, e)
}

func basicAuth(user, pass string) string {
	auth := user + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type metricsStub struct{}

func (m metricsStub) Receive(stats *metrics.EventStats) {}

type fakeProcessor struct {
	*zap.SugaredLogger
}

func (f fakeProcessor) Process(ctx context.Context, ce ce.Event) error {
	f.Infof("received new event: %s", ce.String())
	return nil
}

func (f fakeProcessor) PushMetrics(ctx context.Context, ms metrics.Receiver) {}

func (f fakeProcessor) Shutdown(ctx context.Context) error {
	return nil
}
