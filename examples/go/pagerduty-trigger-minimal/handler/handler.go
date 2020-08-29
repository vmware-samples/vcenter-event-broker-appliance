package function

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	handler "github.com/openfaas/templates-sdk/go-http"
)

const (
	pdConfigPath     = "/var/openfaas/secrets/pdconfig"
	pagerdutyApiPath = "https://events.pagerduty.com/v2/enqueue"
)

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	// Parse the event
	pdObj, err := parseEvent(req.Body)
	if err != nil {
		return respAndErr("parsing event data: %w", err)
	}

	// Read the config
	_, err := readPdConfig(pdConfigPath, &pdObj)
	if err != nil {
		return respAndErr("loading PagerDuty configs: %w", err)
	}

	// Implement business logic
	pdRes, err := pdSendRequest(pagerdutyApiPath, pdObj)
	if err != nil {
		return respAndErr("connecting to PagerDuty API: %w", err)
	}

	// Handle function response
	return handler.Response{
		Body:       pdRes,
		StatusCode: http.StatusOK,
	}, nil
}

func parseEvent(req []byte) (pagerDutyData, error) {
	var ce cloudEvent
	var pd pagerDutyData

	if err := json.Unmarshal(req, &ce); err != nil {
		return pd, err
	}

	if err := isValidEvent(ce); err != nil {
		return pd, err
	}

	pd.Client = "Frankie Go Function"
	pd.ClientURL = ce.Source
	pd.Payload.Summary = ce.Data.FullFormattedMessage
	pd.Payload.Timestamp = ce.Data.CreatedTime
	pd.Payload.Source = ce.Source
	pd.Payload.Severity = "info"
	pd.Payload.Component = ce.Data.Vm.Name
	pd.Payload.Group = ce.Data.Host.Name
	pd.Payload.Class = ce.Subject

	return pd, nil
}

func respAndErr(msg string, err error) (handler.Response, error) {
	wrappedErr := fmt.Errorf(msg, err)

	return handler.Response{
		Body:       []byte(wrappedErr.Error()),
		StatusCode: http.StatusInternalServerError,1
	}, wrappedErr
}

func readPdConfig(path string, pdo *pagerDutyData) (pdConfig, error) {
	var pdc pdConfig

	secret, err := os.Open(path)
	if err != nil {
		return pdc, fmt.Errorf("opening pdconfig secret file: %w", err)
	}

	defer secret.Close()

	jbs, err := ioutil.ReadAll(secret)
	if err != nil {
		return pdc, fmt.Errorf("reading secret: %w", err)
	}

	if err := json.Unmarshal(jbs, &pdc); err != nil {
		return pdc, err
	}

	if err := validatePdConf(pdc); err != nil {
		return pdc, err
	}

	pdo.RoutingKey = pdc.RoutingKey
	pdo.EventAction = pdc.EventAction

	return pdc, nil
}

func pdSendRequest(path string, pdObj pagerDutyData) ([]byte, error) {
	tpCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	clt := &http.Client{
		Transport: tpCfg,
	}

	reqBody, err := json.Marshal(pdObj)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(reqBody))
	if err != nil {
		return []byte{}, err
	}

	res, err := clt.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("executing http request: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("reading http response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return []byte{}, errors.New("http response not successful")
	}

	return body, nil
}
