package metrics

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
)

const (
	// DefaultListenAddress is the default address the http metrics server will listen
	// for requests
	DefaultListenAddress = "0.0.0.0:8080"
	httpTimeout          = time.Second * 5
	endpoint             = "/stats"
)

var (
	eventRouterStats = expvar.NewMap(mapName)
)

// Receiver receives metrics from metric providers
type Receiver interface {
	Receive(stats EventStats)
}

// verify that metrics server implements Receiver
var _ Receiver = (*Server)(nil)

// Server is the implementation of the metrics server
type Server struct {
	http *http.Server
	*log.Logger
}

// NewServer returns an initialized metrics server binding to addr
func NewServer(cfg connection.Config) (*Server, error) {
	var username, password string
	var basicAuth bool

	addr := cfg.Address
	switch cfg.Auth.Method {
	case "basic_auth":
		basicAuth = true
		username = cfg.Auth.Secret["username"]
		password = cfg.Auth.Secret["password"]
	case "none":
		break
	default:
		return nil, errors.Errorf("unsupported authentication method for metrics server: %q", cfg.Auth.Method)
	}

	logger := log.New(os.Stdout, color.Teal("[Metrics Server] "), log.LstdFlags)
	mux := http.NewServeMux()
	switch basicAuth {
	case true:
		mux.Handle(endpoint, withBasicAuth(logger, expvar.Handler(), username, password))
	default:
		mux.Handle(endpoint, expvar.Handler())
	}

	srv := &Server{
		http: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  httpTimeout,
			WriteTimeout: httpTimeout,
		},
		Logger: logger,
	}
	return srv, nil
}

// Run starts the metrics server until the context is cancelled or an error
// occurs. It will collect metrics for the given event streams and processors.
func (s *Server) Run(ctx context.Context, bindAddr string) error {
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		addr := fmt.Sprintf("http://%s%s", bindAddr, endpoint)
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
func withBasicAuth(logger *log.Logger, next http.Handler, u string, p string) http.HandlerFunc {
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
// under the predifined map. The sender is responsible for picking a unique
// Provider name.
func (s *Server) Receive(stats EventStats) {
	eventRouterStats.Set(stats.Provider, &stats)
}
