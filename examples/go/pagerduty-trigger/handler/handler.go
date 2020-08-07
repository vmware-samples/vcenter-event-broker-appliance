package function

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	handler "github.com/openfaas/templates-sdk/go-http"
	toml "github.com/pelletier/go-toml"
)

const (
	pdApiPath    = "https://events.pagerduty.com/v2/enqueue"
	pdConfigPath = "/var/openfaas/secrets/pdconfig"
)

// Handle a function invocation.
func Handle(req handler.Request) (handler.Response, error) {
	ctx := context.Background()

	// Load the PagerDuty config file at every function invocation.
	pdCfg, err := loadConfig(pdConfigPath)
	if err != nil {
		return handlerResponseWithError(
			"load PagerDuty configs: %w",
			http.StatusBadRequest,
			err,
		)
	}

	// Transfer event information data to PagerDuty payload.
	pdPld, err := parseEvent(req.Body, pdCfg)
	if err != nil {
		return handlerResponseWithError(
			"parse event data: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	// Send information to PagerDuty.
	pdRes, err := pdSendRequest(ctx, pdApiPath, pdPld)
	if err != nil {
		return handlerResponseWithError(
			"connect to PagerDuty API: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	// Display PagerDuty response.
	return handler.Response{
		Body:       pdRes,
		StatusCode: http.StatusOK,
	}, nil
}

func handlerResponseWithError(msg string, code int, err error) (handler.Response, error) {
	wrapErr := fmt.Errorf(msg, err)
	log.Println(wrapErr.Error())

	return handler.Response{
		Body:       []byte(wrapErr.Error()),
		StatusCode: code,
	}, wrapErr
}

func loadConfig(path string) (pdConfig, error) {
	var pdc pdConfig

	secret, err := toml.LoadFile(path)
	if err != nil {
		return pdConfig{}, fmt.Errorf("read pdconfig.toml: %w", err)
	}

	if err := secret.Unmarshal(&pdc); err != nil {
		return pdConfig{}, fmt.Errorf("unmarshal pdconfig.toml: %w", err)
	}

	if err := validatePdConf(pdc); err != nil {
		return pdConfig{}, err
	}

	return pdc, nil
}

func validatePdConf(pdc pdConfig) error {
	if pdc.PagerDuty.RoutingKey == "" {
		return errors.New("PagerDuty routing key cannot be empty in pdconfig.toml")
	}

	if pdc.PagerDuty.EventAction == "" {
		return errors.New("PagerDuty event action cannot be empty in pdconfig.toml")
	}

	return nil
}

// parseEvent saves info from the triggering event into a PagerDuty data object.
func parseEvent(req []byte, pdc pdConfig) (pdPayload, error) {
	var ce cloudEvent
	var pd pdPayload

	if err := json.Unmarshal(req, &ce); err != nil {
		return pd, err
	}

	if err := validateEvent(ce); err != nil {
		return pd, err
	}

	pd.RoutingKey = pdc.PagerDuty.RoutingKey
	pd.EventAction = pdc.PagerDuty.EventAction
	pd.Client = "VMware Event Broker Appliance"
	pd.ClientURL = ce.Source
	pd.Payload.Summary = ce.Data.FullFormattedMessage
	pd.Payload.Timestamp = ce.Data.CreatedTime
	pd.Payload.Source = ce.Source
	pd.Payload.Severity = "info"
	pd.Payload.Component = ce.Data.Vm.Name
	pd.Payload.Group = ce.Data.Host.Name
	pd.Payload.Class = ce.Subject
	pd.Payload.CustomDetails.User = ce.Data.UserName
	pd.Payload.CustomDetails.Datacenter = ce.Data.Datacenter
	pd.Payload.CustomDetails.ComputeResource = ce.Data.ComputeResource
	pd.Payload.CustomDetails.Host = ce.Data.Host
	pd.Payload.CustomDetails.VM = ce.Data.Vm

	return pd, nil
}

func validateEvent(event cloudEvent) error {
	var msg string

	switch true {
	case event.Source == "":
		msg = "does not contain Source"
	case event.Subject == "":
		msg = "does not contain Subject"
	case event.Data.FullFormattedMessage == "":
		msg = "does not contain Data.FullFormattedMessage"
	case (event.Data.CreatedTime == time.Time{}):
		msg = "does not contain Data.CreatedTime"
	case event.Data.Vm.Name == "":
		msg = "does not contain Data.Vm.Name"
	case event.Data.Host.Name == "":
		msg = "does not contain Data.Vm.Name"
	}

	if msg != "" {
		return errors.New("invalid event: " + msg)
	}

	return nil
}

func pdSendRequest(ctx context.Context, path string, pdp pdPayload) ([]byte, error) {
	clt := &http.Client{}

	reqBody, err := json.Marshal(pdp)
	if err != nil {
		return []byte{}, err
	}

	res, err := clt.Post(path, "application/json", bytes.NewBuffer(reqBody))
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return []byte{}, fmt.Errorf("execute http request %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read http response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return []byte{}, fmt.Errorf("http response: %d, %+v", res.StatusCode, string(body))
	}

	return body, nil
}
