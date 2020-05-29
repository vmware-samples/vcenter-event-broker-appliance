package processor

import (
	"log"
	"time"

	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
)

// OpenFaaSOption configures the OpenFaaS processor
type OpenFaaSOption func(*openfaasProcessor)

// WithOpenFaaSVerbose enables verbose logging for the OpenFaaS processor
func WithOpenFaaSVerbose(verbose bool) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.verbose = verbose
	}
}

// WithOpenFaaSLogger sets an alternative logger for the OpenFaaS processor
func WithOpenFaaSLogger(logger *log.Logger) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.Logger = logger
	}
}

// WithOpenFaaSDelimiter changes the default topic delimiter (comma-separated
// strings) for the OpenFaaS processor
func WithOpenFaaSDelimiter(delim string) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.topicDelimiter = delim
	}
}

// WithOpenFaaSTimeout changes the default gateway timeout for the OpenFaaS
// processor
func WithOpenFaaSTimeout(timeout time.Duration) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.gatewayTimeout = timeout
	}
}

// WithOpenFaaSRebuildInterval changes the default gateway topic synchronization
// interval for the OpenFaaS processor
func WithOpenFaaSRebuildInterval(interval time.Duration) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.rebuildInterval = interval
	}
}

// WithOpenFaaSResponseHandler sets an alternative response handler for the
// OpenFaaS processor
func WithOpenFaaSResponseHandler(handler ofsdk.ResponseSubscriber) OpenFaaSOption {
	return func(o *openfaasProcessor) {
		o.ResponseSubscriber = handler
	}
}
