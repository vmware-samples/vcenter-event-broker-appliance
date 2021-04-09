// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL     string
	Client         *http.Client
	Credentials    *auth.BasicAuthCredentials
	TopicDelimiter string
}

// getNamespaces gets OpenFaaS namespaces
func (s *FunctionLookupBuilder) getNamespaces() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/namespaces", s.GatewayURL), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	code := res.StatusCode
	if code == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failure against gateway: %s", http.StatusText(code))
	}

	if !(code >= http.StatusOK && code <= 299) {
		return nil, fmt.Errorf("get namespaces unexpected HTTP response: %s", http.StatusText(code))
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	if len(bytesOut) == 0 {
		return nil, errors.New("empty response body")
	}

	var namespaces []string
	err = json.Unmarshal(bytesOut, &namespaces)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal JSON")
	}

	return namespaces, err
}

func (s *FunctionLookupBuilder) getFunctions(namespace string) ([]types.FunctionStatus, error) {
	gateway := fmt.Sprintf("%s/system/functions", s.GatewayURL)
	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway URL: %s", err.Error())
	}
	if len(namespace) > 0 {
		query := gatewayURL.Query()
		query.Set("namespace", namespace)
		gatewayURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, gatewayURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	code := res.StatusCode
	if code == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failure against gateway: %s", http.StatusText(code))
	}

	if !(code >= http.StatusOK && code <= 299) {
		return nil, fmt.Errorf("get functions unexpected HTTP response: %s", http.StatusText(code))
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	var functions []types.FunctionStatus
	err = json.Unmarshal(bytesOut, &functions)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal JSON")
	}

	return functions, nil
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var (
		err error
	)

	namespaces, err := s.getNamespaces()
	if err != nil {
		return map[string][]string{}, err
	}
	serviceMap := make(map[string][]string)

	if len(namespaces) == 0 {
		namespace := ""
		functions, err := s.getFunctions(namespace)
		if err != nil {
			return map[string][]string{}, err
		}
		serviceMap = buildServiceMap(&functions, s.TopicDelimiter, namespace, serviceMap)
	} else {
		for _, namespace := range namespaces {
			functions, err := s.getFunctions(namespace)
			if err != nil {
				return map[string][]string{}, err
			}
			serviceMap = buildServiceMap(&functions, s.TopicDelimiter, namespace, serviceMap)
		}
	}

	return serviceMap, err
}

func buildServiceMap(functions *[]types.FunctionStatus, topicDelimiter, namespace string, serviceMap map[string][]string) map[string][]string {
	for _, function := range *functions {

		if function.Annotations != nil {

			annotations := *function.Annotations

			if topicNames, exist := annotations["topic"]; exist {

				if len(topicDelimiter) > 0 && strings.Count(topicNames, topicDelimiter) > 0 {

					topicSlice := strings.Split(topicNames, topicDelimiter)

					for _, topic := range topicSlice {
						serviceMap = appendServiceMap(topic, function.Name, namespace, serviceMap)
					}
				} else {
					serviceMap = appendServiceMap(topicNames, function.Name, namespace, serviceMap)
				}
			}
		}
	}
	return serviceMap
}

func appendServiceMap(key, function, namespace string, sm map[string][]string) map[string][]string {

	key = strings.TrimSpace(key)

	if len(key) > 0 {

		if sm[key] == nil {
			sm[key] = []string{}
		}
		sep := ""
		if len(namespace) > 0 {
			sep = "."
		}

		functionPath := fmt.Sprintf("%s%s%s", function, sep, namespace)
		sm[key] = append(sm[key], functionPath)
	}

	return sm
}
