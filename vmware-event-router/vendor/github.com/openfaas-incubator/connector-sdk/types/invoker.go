// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

var (
	ErrEmptyMessage = fmt.Errorf("empty message supplied")
)

type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
	Responses     chan InvokerResponse
}

type InvokerResponse struct {
	Context  context.Context
	Body     []byte
	Header   http.Header
	Status   int
	Error    error
	Topic    string
	Function string
}

func NewInvoker(gatewayURL string, client *http.Client, printResponse bool) *Invoker {
	return &Invoker{
		PrintResponse: printResponse,
		Client:        client,
		GatewayURL:    gatewayURL,
		Responses:     make(chan InvokerResponse),
	}
}

// Invoke triggers a function by accessing the API Gateway
func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message []byte) (int, error) {
	return i.InvokeWithContext(context.Background(), topicMap, topic, message)
}

// InvokeWithContext triggers a function by accessing the API Gateway while propagating context
func (i *Invoker) InvokeWithContext(ctx context.Context, topicMap *TopicMap, topic string, message []byte) (int, error) {
	if len(message) == 0 {
		return 0, ErrEmptyMessage
	}

	matchedFunctions := topicMap.Match(topic)
	matched := len(matchedFunctions)
	if matched == 0 {
		return 0, nil
	}

	// parallelize invoking functions
	for _, matchedFunction := range matchedFunctions {
		go func(function string) {
			body, statusCode, header, doErr := i.InvokeFunction(ctx, function, message)

			if doErr != nil {
				i.Responses <- InvokerResponse{
					Context:  ctx,
					Status:   statusCode,
					Header:   header,
					Function: function,
					Topic:    topic,
					Error:    errors.Wrap(doErr, fmt.Sprintf("unable to invoke %s", function)),
				}
				return
			}

			i.Responses <- InvokerResponse{
				Context:  ctx,
				Body:     body,
				Status:   statusCode,
				Header:   header,
				Function: function,
				Topic:    topic,
			}
		}(matchedFunction)
	}

	return matched, nil
}

// InvokeFunction invokes the specified function with the given message body. It
// returns the function response (if any), the http status code, headers and
// error.
func (i *Invoker) InvokeFunction(ctx context.Context, function string, message []byte) ([]byte, int, http.Header, error) {
	fnURL := fmt.Sprintf("%s/%s", i.GatewayURL, function)

	buf := bytes.NewBuffer(message)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fnURL, buf)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, err
	}

	if httpReq.Body != nil {
		defer func() {
			_ = httpReq.Body.Close()
		}()
	}

	res, err := i.Client.Do(httpReq)
	if err != nil {
		return nil, http.StatusInternalServerError, nil, err
	}

	var body []byte
	if res.Body != nil {
		defer func() {
			_ = res.Body.Close()
		}()

		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			// log.Printf("Error reading body")
			return nil, http.StatusInternalServerError, nil, err

		}
	}

	return body, res.StatusCode, res.Header, err
}
