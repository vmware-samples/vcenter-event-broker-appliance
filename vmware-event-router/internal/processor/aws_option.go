package processor

import (
	"log"
	"time"
)

// AWSOption configures the AWS processor
type AWSOption func(*EventBridgeProcessor)

// WithAWSVerbose enables verbose logging for the AWS processor
func WithAWSVerbose(verbose bool) AWSOption {
	return func(aws *EventBridgeProcessor) {
		aws.verbose = verbose
	}
}

// WithAWSLogger sets an alternative logger for the AWS processor
func WithAWSLogger(logger *log.Logger) AWSOption {
	return func(aws *EventBridgeProcessor) {
		aws.Logger = logger
	}
}

// WithAWSResyncInterval configures the interval to sync AWS EventBridge event
// pattern rules
func WithAWSResyncInterval(interval time.Duration) AWSOption {
	return func(aws *EventBridgeProcessor) {
		aws.resyncInterval = interval
	}
}

// WithAWSBatchSize sets the batch size for PutEvents requests
func WithAWSBatchSize(size int) AWSOption {
	return func(aws *EventBridgeProcessor) {
		aws.batchSize = size
	}
}
