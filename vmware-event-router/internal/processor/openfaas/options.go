package openfaas

import (
	"time"

	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
)

// Option configures the OpenFaaS processor
// TODO: change signature to return errors
type Option func(*Processor)

// WithResponseChan sets a custom response channel to use for returning invocation responses
func WithResponseChan(resCh chan ofsdk.InvokerResponse) Option {
	return func(o *Processor) {
		o.respChan = resCh
	}
}

// WithDelimiter changes the default topic delimiter (comma-separated
// strings) for the OpenFaaS processor
func WithDelimiter(delim string) Option {
	return func(o *Processor) {
		o.topicDelimiter = delim
	}
}

// WithTimeout changes the default gateway timeout for the OpenFaaS
// processor
func WithTimeout(timeout time.Duration) Option {
	return func(o *Processor) {
		o.gatewayTimeout = timeout
	}
}

// WithRebuildInterval changes the default gateway topic synchronization
// interval for the OpenFaaS processor
func WithRebuildInterval(interval time.Duration) Option {
	return func(o *Processor) {
		o.rebuildInterval = interval
	}
}

// WithResponseHandler sets an alternative response handler for the
// OpenFaaS processor
func WithResponseHandler(handler ofsdk.ResponseSubscriber) Option {
	return func(o *Processor) {
		o.ResponseSubscriber = handler
	}
}
