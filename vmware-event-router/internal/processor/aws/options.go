package aws

import (
	"time"
)

// Option configures the AWS processor
// TODO: change signature to return errors
type Option func(*EventBridgeProcessor)

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
