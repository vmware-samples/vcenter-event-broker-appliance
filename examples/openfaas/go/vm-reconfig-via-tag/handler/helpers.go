package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	handler "github.com/openfaas/templates-sdk/go-http"
	"github.com/pelletier/go-toml"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
)

// vcConfig represents the toml vcconfig file
type vcConfig struct {
	VCenter struct {
		Server, User, Password string
		Insecure               bool
	}
}

func loadConfig(path string) (*vcConfig, error) {
	var cfg vcConfig

	secret, err := toml.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load vcconfig.toml: %w", err)
	}

	err = secret.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal vcconfig.toml: %w", err)
	}

	err = validateConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("validate vcconfig: %w", err)
	}

	if debug {
		log.Println("vcconfig.toml loaded, unmarshalled, and validated.")
	}

	return &cfg, nil
}

// ValidateConfig ensures the bare minimum of information is in the config file.
func validateConfig(cfg vcConfig) error {
	reqFields := map[string]string{
		"vcenter server":   cfg.VCenter.Server,
		"vcenter user":     cfg.VCenter.User,
		"vcenter password": cfg.VCenter.Password,
	}

	// Multiple fields may be missing, but err on the first encountered.
	for k, v := range reqFields {
		if v == "" {
			return errors.New("required field missing: %q" + k)
		}
	}

	return nil
}

func setVerbosity() bool {
	return os.Getenv("write_debug") == "true"
}

func handlerResponseWithError(msg string, statusCode int, err error) (handler.Response, error) {
	wrapErr := fmt.Errorf(msg, err)
	log.Println(wrapErr.Error())

	return handler.Response{
		Body:       []byte(wrapErr.Error()),
		StatusCode: statusCode,
	}, wrapErr
}

// cloudEvent is a subsection of a Cloud Event.
type cloudEvent struct {
	Data types.Event `json:"data,omitempty"`
}

// eventVmMOR parses the cloud event and gets the triggering VM's managed object
// reference.
func eventVmMOR(req []byte) (*types.ManagedObjectReference, error) {
	var ce cloudEvent
	var mor types.ManagedObjectReference

	err := json.Unmarshal(req, &ce)
	if err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}

	if ce.Data.Vm == nil || ce.Data.Vm.Vm.Value == "" {
		return nil, errors.New("event is not of type VmEvent")
	}

	// Fill information in the request into a govmomi type.
	mor.Type = ce.Data.Vm.Vm.Type
	mor.Value = ce.Data.Vm.Vm.Value

	if debug {
		log.Println("Cloud event parsed and validated.")
	}

	return &mor, nil
}

type vsClient struct {
	govmomi *govmomi.Client
	rest    *rest.Client
	tagMgr  *tags.Manager
}

// newClient connects to vSphere govmomi API
func newClient(ctx context.Context, cfg *vcConfig) (*vsClient, error) {
	u := url.URL{
		Scheme: "https",
		Host:   cfg.VCenter.Server,
		Path:   "sdk",
	}

	u.User = url.UserPassword(cfg.VCenter.User, cfg.VCenter.Password)
	insecure := cfg.VCenter.Insecure

	gc, err := govmomi.NewClient(ctx, &u, insecure)
	if err != nil {
		return nil, fmt.Errorf("connect to vSphere API: %w", err)
	}

	rc := rest.NewClient(gc.Client)
	tm := tags.NewManager(rc)

	vsc := vsClient{
		govmomi: gc,
		rest:    rc,
		tagMgr:  tm,
	}

	err = vsc.rest.Login(ctx, u.User)
	if err != nil {
		return nil, fmt.Errorf("log into REST API: %w", err)
	}

	if debug {
		session, _ := vsc.govmomi.SessionManager.UserSession(ctx)

		log.Printf(
			"New connection to vSphere established. Govmomi SessionID: %s, REST SessionID: %s",
			session.Key,
			vsc.rest.SessionID(),
		)
	}

	return &vsc, nil
}

func authErrReconnect(ctx context.Context, cfg *vcConfig, err error, errMsg string) error {
	if strings.Contains(err.Error(), errMsg) {
		if debug {
			log.Printf("%s. Retry connection to vSphere.\n", errMsg)
		}

		// Set global variable with new connection.
		vsClt, err = newClient(ctx, cfg)
	}

	return err
}

func handleSignal(ctx context.Context, vsc *vsClient) {
	sigCh := make(chan os.Signal, 2)

	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	if debug {
		log.Println("Started signal handling.")
	}

	s := <-sigCh

	if debug {
		log.Printf("Got signal: %v, logging out of vSphere APIs.", s)
	}

	err := vsc.govmomi.Logout(ctx)
	if err != nil {
		log.Printf("log out of Govmomi: %v", err)
	}

	err = vsc.rest.Logout(ctx)
	if err != nil {
		log.Printf("log out of VAPI: %v", err)
	}
}
