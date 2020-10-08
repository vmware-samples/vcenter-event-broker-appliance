package openfaas

import (
	"log"
	"time"

	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
)

// Option configures the OpenFaaS processor
type Option func(*Processor)

// WithVerbose enables verbose logging for the OpenFaaS processor
func WithVerbose(verbose bool) Option {
	return func(o *Processor) {
		o.verbose = verbose
	}
}

// WithLogger sets an alternative logger for the OpenFaaS processor
func WithLogger(logger *log.Logger) Option {
	return func(o *Processor) {
		o.Logger = logger
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
