package processor

import (
	"log"
	"time"
)

// AWSOption configures the AWS processor
type AWSOption func(*awsEventBridgeProcessor)

// WithAWSVerbose enables verbose logging for the AWS processor
func WithAWSVerbose(verbose bool) AWSOption {
	return func(aws *awsEventBridgeProcessor) {
		aws.verbose = verbose
	}
}

// WithAWSLogger sets an alternative logger for the AWS processor
func WithAWSLogger(logger *log.Logger) AWSOption {
	return func(aws *awsEventBridgeProcessor) {
		aws.Logger = logger
	}
}

// WithAWSResyncInterval configures the interval to sync AWS EventBridge event
// pattern rules
func WithAWSResyncInterval(interval time.Duration) AWSOption {
	return func(aws *awsEventBridgeProcessor) {
		aws.resyncInterval = interval
	}
}

// WithAWSBatchSize sets the batch size for PutEvents requests
func WithAWSBatchSize(size int) AWSOption {
	return func(aws *awsEventBridgeProcessor) {
		aws.batchSize = size
	}
}
