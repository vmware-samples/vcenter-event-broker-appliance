package aws

import (
	"log"
	"time"
)

// Option configures the AWS processor
type Option func(*EventBridgeProcessor)

// WithVerbose enables verbose logging for the AWS processor
func WithVerbose(verbose bool) Option {
	return func(aws *EventBridgeProcessor) {
		aws.verbose = verbose
	}
}

// WithLogger sets an alternative logger for the AWS processor
func WithLogger(logger *log.Logger) Option {
	return func(aws *EventBridgeProcessor) {
		aws.Logger = logger
	}
}

// WithResyncInterval configures the interval to sync AWS EventBridge event
// pattern rules
func WithResyncInterval(interval time.Duration) Option {
	return func(aws *EventBridgeProcessor) {
		aws.resyncInterval = interval
	}
}

// WithBatchSize sets the batch size for PutEvents requests
func WithBatchSize(size int) Option {
	return func(aws *EventBridgeProcessor) {
		aws.batchSize = size
	}
}
