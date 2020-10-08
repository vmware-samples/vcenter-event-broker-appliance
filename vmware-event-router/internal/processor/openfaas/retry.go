package openfaas

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync/atomic"

	"github.com/avast/retry-go"
	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/pkg/errors"
)

// isRetryable provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func isRetryable(ctx context.Context, code int, err error) (bool, error) {
	// source: https://github.com/hashicorp/go-retryablehttp/blob/master/client.go
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe := regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe := regexp.MustCompile(`unsupported protocol scheme`)

	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, nil
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if code == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if code == 0 || (code >= 500 && code != 501) {
		return true, nil
	}

	return false, nil
}

// retryFunc returns a function with internal retry logic based on the given
// initial response. Retries, if any, will be performed with invoker and message
// (to perform retries)
func retryFunc(ctx context.Context, res types.InvokerResponse, invoker invokeFn, msg []byte, counter *int32) func() error {
	var (
		resStatus = res.Status
		resError  = res.Error
		resMsg    []byte
	)

	return func() error {
		retryable, err := isRetryable(ctx, resStatus, resError)
		if err != nil || !retryable {
			return retry.Unrecoverable(errors.Errorf("could not invoke function %q on topic %q: %v", res.Function, res.Topic, err))
		}

		// retry
		atomic.AddInt32(counter, 1)
		resMsg, resStatus, _, resError = invoker(ctx, res.Function, msg)
		if !isSuccessful(resStatus, resError) {
			return fmt.Errorf("function %q on topic %q returned non successful status code %d: %q", res.Function, res.Topic, resStatus, string(resMsg))
		}
		return nil
	}
}

// isSuccessful returns true if no error has occurred and when the HTTP status
// code is 2xx
func isSuccessful(status int, err error) bool {
	return err == nil && (status >= http.StatusOK && status <= 299)
}
