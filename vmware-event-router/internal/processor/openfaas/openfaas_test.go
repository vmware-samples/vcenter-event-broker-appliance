//go:build unit

package openfaas

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/avast/retry-go"
	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"github.com/pkg/errors"
	"go.uber.org/zap/zaptest"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
)

// counter is concurrency-safe
type counter struct {
	sync.RWMutex
	count int
}

func (c *counter) get() int {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func (c *counter) increment() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

func Test_waitForAll(t *testing.T) {
	// simulate numWait functions to wait for after their invocation
	const numWait = 5
	var c counter

	ctx := context.Background()
	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// this func never fails
	wfOk := func(ctx context.Context) error { return nil }

	// this func fails based on a mod counter%numWait condition
	wfFail := func(ctx context.Context) error {
		if c.get()%numWait >= 2 {
			return errors.New("failure occurred")
		}
		c.increment()
		return nil
	}

	// this func fails with context cancelled err
	wfFailCtx := func(ctx context.Context) error {
		cancel()
		return ctx.Err()
	}

	type args struct {
		ctx   context.Context
		waitN int
		fn    waitFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "All waitFn succeed",
			args: args{
				ctx:   ctx,
				waitN: numWait,
				fn:    wfOk,
			},
			wantErr: false,
		},
		{
			name: "Partial failure",
			args: args{
				ctx:   ctx,
				waitN: numWait,
				fn:    wfFail,
			},
			wantErr: true,
		},
		{
			name: "Context cancelled",
			args: args{
				ctx:   ctxCancel,
				waitN: numWait,
				fn:    wfFailCtx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := waitForAll(tt.args.ctx, tt.args.waitN, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("waitForAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_waitForOne(t *testing.T) {
	var (
		log         = zaptest.NewLogger(t).Sugar()
		ctx         = context.Background()
		numTests    = 5 // update when number of tests changes
		retryCount  = 0 // track retries per function
		maxRetries  = 3
		retryCounts = make([]int, numTests) // desired number of retries per test
		responses   = make([]ofsdk.InvokerResponse, numTests)
		resChan     = make(chan ofsdk.InvokerResponse, numTests)
		retryOpts   = []retry.Option{retry.Attempts(uint(maxRetries))}
	)

	// this retry func never fails
	okFunc := func(ctx context.Context, fn string, message []byte) ([]byte, int, http.Header, error) {
		// when this function is called it is a retry attempt
		retryCount++
		return []byte("OK"), 200, http.Header{}, nil
	}

	// this retry func always fails
	failFunc := func(ctx context.Context, fn string, message []byte) ([]byte, int, http.Header, error) {
		// when this function is called it is a retry attempt
		retryCount++
		return []byte("Error"), 500, http.Header{}, nil
	}

	// this retry func fails with redirect error
	failWithErrFunc := func(ctx context.Context, fn string, message []byte) ([]byte, int, http.Header, error) {
		// when this function is called it is a retry attempt
		retryCount++

		// this error won't be retried by the internal mechanism
		return []byte("Error"), 500, http.Header{}, &url.Error{
			Op:  "",
			URL: "",
			Err: errors.New("stopped after 1 redirects"),
		}
	}

	// this retry func succeeds on second retry
	retryableFunc := func(ctx context.Context, fn string, message []byte) ([]byte, int, http.Header, error) {
		// when this function is called it is a retry attempt
		retryCount++
		if retryCount == 2 {
			return []byte("OK"), 200, http.Header{}, nil
		}
		return []byte("Error"), 500, http.Header{}, nil
	}

	// create timeout ctx and immediately force timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	cancel()

	type args struct {
		ctx       context.Context
		resCh     <-chan ofsdk.InvokerResponse
		invoker   invokeFunc
		retryMsg  []byte
		log       logger.Logger
		retryOpts []retry.Option
	}
	tests := []struct {
		name           string
		args           args
		firstResponse  ofsdk.InvokerResponse // response from the first invocation (before retries)
		wantRetryCount int                   // track retried attempts
		wantErr        bool
	}{
		{
			name: "First invocation is immediately successful (no retries)",
			args: args{
				ctx:       ctx,
				resCh:     resChan,
				invoker:   okFunc,
				retryMsg:  []byte("should not retry"),
				log:       log,
				retryOpts: retryOpts,
			},
			firstResponse: ofsdk.InvokerResponse{
				Body:     []byte("OK"),
				Status:   200,
				Error:    nil,
				Topic:    "test-topic",
				Function: "test-function-http_200",
			},
			wantRetryCount: 0,
			wantErr:        false,
		},
		{
			name: "Invocation always fails (stop after max retries)",
			args: args{
				ctx:       ctx,
				resCh:     resChan,
				invoker:   failFunc,
				retryMsg:  []byte("should retry"),
				log:       log,
				retryOpts: retryOpts,
			},
			firstResponse: ofsdk.InvokerResponse{
				Body:     []byte("OK"),
				Status:   500,
				Error:    nil,
				Topic:    "test-topic",
				Function: "test-function-http_500",
			},
			wantRetryCount: 3,
			wantErr:        false,
		},
		{
			name: "Retry func fails with error (stop after first retry)",
			args: args{
				ctx:       ctx,
				resCh:     resChan,
				invoker:   failWithErrFunc,
				retryMsg:  []byte("should retry"),
				log:       log,
				retryOpts: retryOpts,
			},
			firstResponse: ofsdk.InvokerResponse{
				Body:     []byte("OK"),
				Status:   500,
				Error:    nil,
				Topic:    "test-topic",
				Function: "test-function-http_500",
			},
			wantRetryCount: 1,
			wantErr:        false,
		},
		{
			name: "Retry func succeeds (stop after second retry)",
			args: args{
				ctx:       ctx,
				resCh:     resChan,
				invoker:   retryableFunc,
				retryMsg:  []byte("should retry"),
				log:       log,
				retryOpts: retryOpts,
			},
			firstResponse: ofsdk.InvokerResponse{
				Body:     []byte("OK"),
				Status:   500,
				Error:    nil,
				Topic:    "test-topic",
				Function: "test-function-http_500",
			},
			wantRetryCount: 2,
			wantErr:        false,
		},
		{
			name: "First invocation failed and meanwhile context timed out (no retry)",
			args: args{
				ctx:       timeoutCtx,
				resCh:     resChan,
				invoker:   okFunc,
				retryMsg:  []byte("should not retry"),
				log:       log,
				retryOpts: retryOpts,
			},
			firstResponse: ofsdk.InvokerResponse{
				Body:     []byte("OK"),
				Status:   500,
				Error:    nil,
				Topic:    "test-topic",
				Function: "test-function-http_500",
			},
			wantRetryCount: 0,
			wantErr:        false,
		},
	}
	for idx, tt := range tests {
		// set expectations
		retryCounts[idx] = tt.wantRetryCount
		responses[idx] = tt.firstResponse
		resChan <- responses[idx]

		// reset counter
		retryCount = 0

		t.Run(tt.name, func(t *testing.T) {
			wf := waitForOne(tt.args.resCh, tt.args.invoker, tt.args.retryMsg, tt.args.log, tt.args.retryOpts...)

			// run function
			if err := wf(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("waitForOne() error = %v, wantErr %v", err, tt.wantErr)
			}

			if retryCount != tt.wantRetryCount {
				t.Errorf("waitForOne() retries = %v, wantRetries %v", retryCount, tt.wantRetryCount)
			}
		})
	}
}

func Test_waitForOneProcStopped(t *testing.T) {
	var (
		resChan = make(chan ofsdk.InvokerResponse)
		log     = zaptest.NewLogger(t).Sugar()
	)

	type args struct {
		resCh     <-chan ofsdk.InvokerResponse
		invoker   invokeFunc
		retryMsg  []byte
		log       logger.Logger
		retryOpts []retry.Option
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Processor stopped",
			args: args{
				resCh:     resChan,
				invoker:   nil,
				retryMsg:  nil,
				log:       log,
				retryOpts: nil,
			},
			wantErr: ErrStopped,
		},
	}

	close(resChan)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := waitForOne(tt.args.resCh, tt.args.invoker, tt.args.retryMsg, tt.args.log, tt.args.retryOpts...)

			// run function
			if err := wf(context.Background()); err != tt.wantErr {
				t.Errorf("waitForOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
