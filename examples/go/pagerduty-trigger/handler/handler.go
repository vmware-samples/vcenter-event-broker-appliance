package function

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	handler "github.com/openfaas/templates-sdk/go-http"
	toml "github.com/pelletier/go-toml"
)

const (
	pagerdutyApiPath = "https://events.pagerduty.com/v2/enqueue"
	pdConfigPath     = "/var/openfaas/secrets/pdconfig"

	// Background colors for debug messages.
	bgcHeader    = "\033[95m"
	bgcOkBlue    = "\033[94m"
	bgcOkGreen   = "\033[92m"
	bgcWarning   = "\033[93m"
	bgcFail      = "\033[91m"
	bgcEndC      = "\033[0m"
	bgcBold      = "\033[1m"
	bgcUnderline = "\033[4m"
)

func readEnvVars(pdc *pdConfig) {
	// Stay with default of false unless true is set.
	if os.Getenv("write_debug") == "true" {
		logDebug(pdc, "warn", "debug has been enabled for this function. Sensitive information could be printed to sysout")
	}

	// Stay with default of false unless true is set.
	if os.Getenv("tls_insecure") == "true" {
		pdc.TlsInsecure = false
		logDebug(pdc, "warn", "connections are not TLS secured")
	}
}

// logDebug sends formatted message to stdOut if debug is enabled.
func logDebug(pdc *pdConfig, level string, msg string) {
	if pdc.Debug {
		switch level {
		case "error":
			log.Printf("\n%sError, %s.%s\n", bgcFail, msg, bgcEndC)
		case "warn":
			log.Printf("\n%sWarning, %s.%s\n", bgcWarning, msg, bgcEndC)
		default:
			log.Printf("\n%s\n", msg)
		}
	}
}

// Handle a function invocation.
func Handle(req handler.Request) (handler.Response, error) {
	ctx := context.Background()

	// Load the Pager Duty config file at every function invocation.
	pdCfg, err := loadPdConfig(pdConfigPath)
	if err != nil {
		wrapErr := fmt.Errorf("loading PagerDuty configs: %w", err)
		log.Println(wrapErr.Error())

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	// Transfer event information data to pagerduty data object.
	pdObj, err := parseEvent(req.Body, pdCfg)
	if err != nil {
		wrapErr := fmt.Errorf("parsing event data: %w", err)
		log.Println(wrapErr.Error())

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	// Send information to pagerduty.
	pdRes, err := pdSendRequest(ctx, pdCfg, pagerdutyApiPath, pdObj)
	if err != nil {
		wrapErr := fmt.Errorf("connecting to PagerDuty API: %w", err)
		log.Println(wrapErr.Error())

		return handler.Response{
			Body:       []byte(wrapErr.Error()),
			StatusCode: http.StatusInternalServerError,
		}, wrapErr
	}

	// Display pagerduty response.
	return handler.Response{
		Body:       pdRes,
		StatusCode: http.StatusOK,
	}, nil
}

func loadPdConfig(path string) (*pdConfig, error) {
	var pdc pdConfig

	readEnvVars(&pdc)

	secret, err := toml.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pdconfig secret file: %w", err)
	}

	if err := secret.Unmarshal(&pdc); err != nil {
		return nil, fmt.Errorf("unmarshalling pdconfig secret file: %w", err)
	}

	if err := validatePdConf(&pdc); err != nil {
		return nil, err
	}

	return &pdc, nil
}

func validatePdConf(pdc *pdConfig) error {
	if pdc.PagerDuty.RoutingKey == "" {
		return errors.New("PagerDuty routing key cannot be empty")
	}

	if pdc.PagerDuty.EventAction == "" {
		return errors.New("PagerDuty event action cannot be empty")
	}

	return nil
}

// parseEvent saves info from the triggering event into a PagerDuty data object.
func parseEvent(req []byte, pdc *pdConfig) (pagerDutyData, error) {
	var ce cloudEvent
	var pd pagerDutyData

	if err := json.Unmarshal(req, &ce); err != nil {
		return pd, err
	}

	if err := isValidEventResp(ce); err != nil {
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

func isValidEventResp(event cloudEvent) error {
	if event.Data.Vm == nil || event.Data.Vm.Vm.Value == "" {
		return errors.New("empty managed object reference")
	}

	return nil
}

func pdSendRequest(ctx context.Context, pdc *pdConfig, path string, pdObj pagerDutyData) ([]byte, error) {
	tpCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: pdc.TlsInsecure,
		},
	}

	clt := &http.Client{
		Transport: tpCfg,
	}

	reqBody, err := json.Marshal(pdObj)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", path, bytes.NewBuffer(reqBody))
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create http request due to %w", err)
	}

	req.Header.Add("token", pdc.PagerDuty.RoutingKey)

	res, err := clt.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("executing http request %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("reading http response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return []byte{}, fmt.Errorf("http response: %d, %+v", res.StatusCode, string(body))
	}

	return body, nil
}
