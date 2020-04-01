package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	handler "github.com/openfaas-incubator/go-function-sdk"
	"github.com/pelletier/go-toml"
	"github.com/vmware/govmomi/vim25/types"
)

const cfgPath = "/var/openfaas/secrets/vcconfig"

// vcConfig represents the toml vcconfig file
type vcConfig struct {
	VCenter struct {
		Server   string
		User     string
		Password string
		Insecure bool
	}
	Tag struct {
		URN    string
		Action string
	}
}

// Incoming is a subsection of a Cloud Event.
type incoming struct {
	Data types.Event `json:"data,omitempty"`
}

var (
	lock   sync.Mutex // Lock protects client.
	client *vsClient  // Client persists vSphere connection.
	once   sync.Once  // For handleSignal() to be called once.
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	ctx := context.Background()

	// Load config every time, to ensure the most updated version is used.
	cfg, err := loadTomlCfg(cfgPath)
	if err != nil {
		wrapErr := fmt.Errorf("loading of vcconfig failed: %w", err)
		log.Println(wrapErr.Error())

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	// Connect to vSphere govmomi API once and persist connection with global variable.
	err = vsConnect(ctx, cfg)
	if err != nil {
		wrapErr := fmt.Errorf("connect to vSphere failed: %w", err)

		if debug() {
			log.Println(wrapErr)
		}

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	once.Do(func() {
		// Set up os signal handling to log out of vSphere.
		go handleSignal(ctx)
	})

	// Retrieve the Managed Object Reference from the event.
	moRef, err := parseEventMoRef(req.Body)
	if err != nil {
		wrapErr := fmt.Errorf("retrieve managed reference object failed: %w", err)

		if debug() {
			log.Println(wrapErr)
		}

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusBadRequest,
		}, wrapErr
	}

	err = client.moTag(ctx, *moRef, cfg.Tag.URN)
	if err != nil {
		wrapErr := fmt.Errorf("tagging managed reference object failed: %w", err)

		if debug() {
			log.Println(wrapErr)
		}

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	message := fmt.Sprintf("%v was tagged with %v", moRef.Value, cfg.Tag.URN)
	log.Println(message)

	return handler.Response{
		Body:       []byte(message),
		StatusCode: http.StatusOK,
	}, nil
}

// vsConnect connects to vSphere govmomi API using information from vcconfig.toml.
func vsConnect(ctx context.Context, cfg *vcConfig) error {
	lock.Lock()
	defer lock.Unlock()

	if client == nil {
		u := url.URL{
			Scheme: "https",
			Host:   cfg.VCenter.Server,
			Path:   "sdk",
		}
		u.User = url.UserPassword(cfg.VCenter.User, cfg.VCenter.Password)
		insecure := cfg.VCenter.Insecure

		if debug() {
			log.Println("connect to vSphere")
		}

		c, err := newClient(ctx, u, insecure)
		if err != nil {
			return fmt.Errorf("connection to vSphere API failed: %w", err)
		}

		// Set global variable to persist connection.
		client = c
	}

	return nil
}

func loadTomlCfg(path string) (*vcConfig, error) {
	var cfg vcConfig

	secret, err := toml.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load vcconfig.toml: %w", err)
	}

	err = secret.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal vcconfig.toml: %w", err)
	}

	err = validateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("insufficient information in vcconfig.toml: %w", err)
	}

	return &cfg, nil
}

// ValidateConfig ensures the bare minimum of information is in the config file.
func validateConfig(cfg vcConfig) error {
	reqFields := map[string]string{
		"vcenter server":   cfg.VCenter.Server,
		"vcenter user":     cfg.VCenter.User,
		"vcenter password": cfg.VCenter.Password,
		"tag URN":          cfg.Tag.URN,
		"tag action":       cfg.Tag.Action,
	}

	// Multiple fields may be missing, but err on the first encountered.
	for k, v := range reqFields {
		if v == "" {
			return errors.New("required field(s) missing, including " + k)
		}
	}

	return nil
}

// Debug determines verbose logging
func debug() bool {
	verbose := os.Getenv("write_debug")

	if verbose == "true" {
		return true
	}

	return false
}

func parseEventMoRef(req []byte) (*types.ManagedObjectReference, error) {
	var event incoming
	var moRef types.ManagedObjectReference

	err := json.Unmarshal(req, &event)
	if err != nil {
		return nil, fmt.Errorf("parsing of request failed: %w", err)
	}

	if event.Data.Vm == nil || event.Data.Vm.Vm.Value == "" {
		return nil, errors.New("empty managed reference object")
	}

	// Fill information in the request into a govmomi type.
	moRef.Type = event.Data.Vm.Vm.Type
	moRef.Value = event.Data.Vm.Vm.Value

	return &moRef, nil
}

func handleSignal(ctx context.Context) {
	var sigCh = make(chan os.Signal, 2)

	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	s := <-sigCh
	verbose := debug()

	if verbose {
		log.Printf("got signal: %v, log out of vSphere", s)
	}

	err := client.logout(ctx)
	if verbose {
		if err != nil {
			log.Printf("vSphere logout failed: %v", err)
			return
		}
		log.Println("logged out of govmomi and rest APIs")
	}
}
