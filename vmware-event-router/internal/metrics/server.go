package metrics

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
)

const (
	// DefaultListenAddress is the default address the http metrics server will listen
	// for requests
	httpTimeout = time.Second * 5
	endpoint    = "/stats"
)

var (
	eventRouterStats = expvar.NewMap(mapName)
)

// Receiver receives metrics from metric providers
type Receiver interface {
	Receive(stats *EventStats)
}

// verify that metrics server implements Receiver
var _ Receiver = (*Server)(nil)

// Server is the implementation of the metrics server
type Server struct {
	http *http.Server
	*log.Logger
}

// NewServer returns an initialized metrics server binding to addr
func NewServer(cfg *config.MetricsProviderConfigDefault) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("no metrics server configuration found")
	}

	logger := log.New(os.Stdout, color.Teal("[Metrics Server] "), log.LstdFlags)
	basicAuth := true

	if cfg.Auth == nil || cfg.Auth.BasicAuth == nil {
		logger.Print("no credentials found, disabling authentication for metrics server")
		basicAuth = false
	}

	mux := http.NewServeMux()

	switch basicAuth {
	case true:
		mux.Handle(endpoint, withBasicAuth(logger, expvar.Handler(), cfg.Auth.BasicAuth.Username, cfg.Auth.BasicAuth.Password))
	default:
		mux.Handle(endpoint, expvar.Handler())
	}

	err := validateAddress(cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrap(err, "could not validate bind address")
	}

	srv := &Server{
		http: &http.Server{
			Addr:         cfg.BindAddress,
			Handler:      mux,
			ReadTimeout:  httpTimeout,
			WriteTimeout: httpTimeout,
		},
		Logger: logger,
	}

	return srv, nil
}

// validateAddress validates the given address and will return an error if the
// format is not <IP>:<PORT>
func validateAddress(address string) error {
	// TODO: this list is not extensive and needs to be changed once we allow DNS
	// names for external metrics endpoints
	const invalidChars = `abcdefghijklmnopqrstuvwxyz/\ `

	address = strings.ToLower(address)
	if strings.ContainsAny(address, invalidChars) {
		return errors.New("invalid character detected (required format: <IP>:<PORT>)")
	}

	// 	check if port if specified
	if !strings.Contains(address, ":") {
		return errors.New("no port specified")
	}

	h, p, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	if h == "" {
		return errors.New("no IP listen address specified")
	}

	if p == "" {
		return errors.New("no port specified")
	}

	return nil
}

// Run starts the metrics server until the context is cancelled or an error
// occurs. It will collect metrics for the given event streams and processors.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		addr := fmt.Sprintf("http://%s%s", s.http.Addr, endpoint)
		s.Printf("starting metrics server and listening on %q", addr)

		err := s.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// continuously update the http stats endpoint
	go func() {
		s.publish(ctx)
	}()

	select {
	case <-ctx.Done():
		err := s.http.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			return errors.Wrap(err, "could not shutdown metrics server gracefully")
		}
	case err := <-errCh:
		return errors.Wrap(err, "could not run metrics server")
	}

	return nil
}

// withBasicAuth enforces basic auth as a middleware for the given username and
// password
func withBasicAuth(logger *log.Logger, next http.Handler, u, p string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		if !ok || !(p == password && u == user) {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("invalid credentials"))

			if err != nil {
				logger.Printf("could not write http response: %v", err)
			}

			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Server) publish(ctx context.Context) {
	var (
		numberOfSecondsRunning = expvar.NewInt("system.numberOfSeconds") // uptime in sec
		programName            = expvar.NewString("system.programName")
		lastLoad               = expvar.NewFloat("system.lastLoad")
	)

	expvar.Publish("system.allLoad", expvar.Func(allLoadAvg))
	programName.Set(os.Args[0])

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			numberOfSecondsRunning.Add(1)
			lastLoad.Set(loadAvg(0))
		}
	}
}

// Receive receives metrics from event streams and processors and exposes them
// under the predefined map. The sender is responsible for picking a unique
// Provider name.
func (s *Server) Receive(stats *EventStats) {
	eventRouterStats.Set(stats.Provider, stats)
}
