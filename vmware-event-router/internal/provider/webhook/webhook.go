package webhook

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/pkg/errors"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/util"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

const (
	// webhook endpoint settings
	defaultPath    = "webhook"
	allowedOrigins = "*"
	allowedMethod  = "POST"

	// TODO: make configurable
	pollConcurrency = 1 // goroutines polling in receive

	// TODO: currently not implemented in CE SDK
	// TODO: make configurable
	allowedRate = 1000
)

var (
	// ErrInvalidPath is returned on an invalid webhook endpoint path
	ErrInvalidPath = errors.New("invalid webhook endpoint path")
)

// Server is a webhook event provider
type Server struct {
	ceclient ce.Client
	listener net.Listener // holds net.Listener
	logger.Logger

	sync.RWMutex
	stats metrics.EventStats
}

// NewServer returns a webhook event provider
func NewServer(ctx context.Context, cfg *config.ProviderConfigWebhook, ms metrics.Receiver, log logger.Logger, opts ...Option) (*Server, error) {
	var srv Server
	if err := util.ValidateAddress(cfg.BindAddress); err != nil {
		return nil, errors.Wrap(err, "invalid webhook config")
	}

	path, err := validatePath(cfg.Path)
	if err != nil {
		return nil, errors.Wrap(err, "invalid webhook config")
	}

	srv.Logger = log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		prov := strings.ToUpper(string(config.ProviderWebhook))
		srv.Logger = zapSugared.Named(fmt.Sprintf("[%s]", prov))
		ctx = logging.WithLogger(ctx, srv.Logger.(*zap.SugaredLogger))
	}

	l, err := net.Listen("tcp", cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrap(err, "start listener")
	}

	// default client options
	ceOpts := []cehttp.Option{
		ce.WithListener(l),
		ce.WithPath("/" + path),
		ce.WithDefaultOptionsHandlerFunc([]string{allowedMethod}, allowedRate, []string{allowedOrigins}, true),
	}

	// middleware options (executed in reverse order)
	mwOpts := []cehttp.Option{
		ce.WithMiddleware(func(next http.Handler) http.Handler {
			return withLogger(srv.Logger, next)
		}),
	}

	if cfg.Auth == nil || cfg.Auth.BasicAuth == nil {
		srv.Warnf("disabling basic auth: no authentication data provided")
	} else {
		srv.Info("enabling endpoint authentication with basic auth")

		// hack: prepend the auth middleware so other mw are run before (mw executed in reverse order)
		// TODO: optimize for less allocations
		mwOpts = append([]cehttp.Option{ce.WithMiddleware(func(next http.Handler) http.Handler {
			return withBasicAuth(ctx, next, cfg.Auth.BasicAuth.Username, cfg.Auth.BasicAuth.Password)
		})}, mwOpts...)
	}

	ceOpts = append(ceOpts, mwOpts...)
	p, err := ce.NewHTTP(ceOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "create cloud event protocol")
	}

	client, err := ce.NewClient(p, ceclient.WithPollGoroutines(pollConcurrency))
	if err != nil {
		return nil, errors.Wrap(err, "create cloud event client")
	}

	srv.ceclient = client
	srv.listener = l
	srv.stats = metrics.EventStats{
		Provider:    string(config.ProviderWebhook),
		Type:        config.EventProvider,
		Address:     cfg.BindAddress,
		Started:     time.Now().UTC(),
		EventsTotal: new(int),
		EventsErr:   new(int),
		EventsSec:   new(float64),
	}

	// apply options (use defaults otherwise)
	for _, opt := range opts {
		opt(&srv)
	}

	srv.Debugw("cloud event protocol configured", "port", p.GetListeningPort(), "path", p.GetPath())

	go srv.PushMetrics(ctx, ms)

	return &srv, nil
}

// validatePath removes any leading and trailing slashes and then validates the
// given webhook endpoint path. If the path is empty, the default path will be
// returned. Root path "/" is not allowed
func validatePath(path string) (string, error) {
	if path == "" {
		return defaultPath, nil
	}

	// root path is not allowed
	if path == "/" {
		return "", ErrInvalidPath
	}

	path = strings.TrimLeft(path, "/")
	path = strings.TrimRight(path, "/")
	path = url.PathEscape(path)
	path = strings.ToLower(path)

	return path, nil
}

// withBasicAuth enforces basic auth as a middleware for the given username and
// password
func withBasicAuth(_ context.Context, next http.Handler, u, p string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			// reduce brute-force guessing attacks with constant-time comparisons
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(u))
			expectedPasswordHash := sha256.Sum256([]byte(p))

			usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
			passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// withLogger logs the incoming http request in DEBUG level
func withLogger(logger logger.Logger, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debugw("incoming http request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr, "headers", r.Header)
		next.ServeHTTP(w, r)
	}
}

// Address returns the listener address and port, e.g. "10.0.0.1:8080"
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// PushMetrics pushes metrics to the configured metrics receiver
func (s *Server) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Lock()
			eventsSec := math.Round((float64(*s.stats.EventsTotal)/time.Since(s.stats.Started).Seconds())*100) / 100 // 0.2f syntax
			s.stats.EventsSec = &eventsSec
			ms.Receive(&s.stats)
			s.Unlock()
		}
	}
}

// Stream starts the webhook event provider invoking the specified processor for
// every incoming valid CloudEvent. Stream will return when the given context is cancelled.
func (s *Server) Stream(ctx context.Context, proc processor.Processor) error {
	s.Info("starting webhook server")
	if err := s.ceclient.StartReceiver(ctx, s.processEvent(proc)); err != nil {
		return errors.Wrap(err, "start webhook server")
	}
	return nil
}

// receiveFunc is a valid signature for a CloudEvent client receiver handler
type receiveFunc func(ctx context.Context, e ce.Event) ce.Result

// processEvent injects a processor into a receiveFunc
func (s *Server) processEvent(p processor.Processor) receiveFunc {
	return func(ctx context.Context, e ce.Event) ce.Result {
		err := p.Process(ctx, e)

		s.Lock()
		defer s.Unlock()

		*s.stats.EventsTotal++
		if err != nil {
			*s.stats.EventsErr++
		}
		return err
	}
}

// Shutdown is a no-op. The webhook server will shut down when the context in
// Stream() is cancelled.
func (s *Server) Shutdown(_ context.Context) error {
	return nil
}
