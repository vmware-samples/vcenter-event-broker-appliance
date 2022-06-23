package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	cectx "github.com/cloudevents/sdk-go/v2/context"
	"github.com/embano1/vsphere/logger"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	httpTimeout     = time.Second * 5
	sourceFormat    = "/%s"                     // /K_SERVICE
	eventTypeFormat = "com.vmware.harbor.%s.v0" // com.vmware.harbor.pull_artifact.v0

	// secrets
	userFileKey     = "username"
	passwordFileKey = "password"
)

var (
	retries    = 3
	retryDelay = time.Millisecond * 200
)

func run(ctx context.Context, cfg config) error {
	log := logger.Get(ctx)

	client, err := ce.NewClientHTTP(ce.WithTarget(cfg.Sink))
	if err != nil {
		return fmt.Errorf("create cloudevent client: %w", err)
	}

	var auth bool
	if path := os.Getenv("WEBHOOK_SECRET_PATH"); path != "" {
		auth = true
	}

	handler := eventHandler(ctx, client)
	if auth {
		user, err := readKey(userFileKey, cfg.SecretPath)
		if err != nil {
			return fmt.Errorf("read secret key %q: %w", userFileKey, err)
		}

		pass, err := readKey(passwordFileKey, cfg.SecretPath)
		if err != nil {
			return fmt.Errorf("read secret key %q: %w", passwordFileKey, err)
		}

		handler = withBasicAuth(ctx, eventHandler(ctx, client), user, pass)
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, handler)

	address := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	srv := http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-egCtx.Done()
		log.Info("shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Warn("could not gracefully shutdown http server")
		}
		return nil
	})

	log.Info("starting http server",
		zap.String("address", address),
		zap.String("path", cfg.Path),
		zap.String("sink", cfg.Sink),
		zap.Bool("basic_auth", auth),
	)

	eg.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("run http server: %w", err)
		}
		return nil
	})

	return eg.Wait()
}

// harbor webhook event handler
func eventHandler(ctx context.Context, client ce.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		// TODO (@mgasch): support inbound rate limiting

		log := logger.Get(ctx)
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("read body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var event model.Payload
		if err = json.Unmarshal(b, &event); err != nil {
			log.Error("could not decode harbor notification event", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		id := uuid.New().String()
		log = log.With(zap.String("eventID", id))

		log.Debug("received request", zap.String("request", string(b)))

		e := ce.NewEvent()
		e.SetID(id)
		e.SetSource(fmt.Sprintf(sourceFormat, os.Getenv("K_SERVICE")))
		e.SetSubject(event.Operator) // might be empty

		// sanity check
		if event.Type == "" {
			log.Error("harbor event type must not be empty", zap.String("type", event.Type))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		t := strings.ToLower(event.Type)
		e.SetType(fmt.Sprintf(eventTypeFormat, t))

		ts := time.Unix(event.OccurAt, 0)
		e.SetTime(ts)

		if err = e.SetData(ce.ApplicationJSON, event); err != nil {
			log.Error("could not set cloudevent data", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx = cectx.WithRetriesExponentialBackoff(ctx, retryDelay, retries)
		if err = client.Send(ctx, e); ce.IsUndelivered(err) || ce.IsNACK(err) {
			log.Error("could not send cloudevent", zap.Error(err), zap.String("event", e.String()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		log.Debug("successfully sent cloudevent", zap.Any("event", e))
	})
}

// withBasicAuth enforces basic auth as a middleware for the given username and
// password
func withBasicAuth(ctx context.Context, next http.Handler, u, p string) http.Handler {
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

		logger.Get(ctx).Debug("rejecting incoming request: user not authenticated")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// readKey reads the file from the secret path
func readKey(key string, path string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(path, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
