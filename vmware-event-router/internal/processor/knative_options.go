package processor

// KnativeOption configures the knative processor
type KnativeOption func(*knativeProcessor)

// WithKnativeVerbose enables verbose logging for the Knative processor
func WithKnativeVerbose(verbose bool) KnativeOption {
	return func(knative *knativeProcessor) {
		knative.verbose = verbose
	}
}

// WithKnativeRetry enables Knative processor to Send events, retrying in case of a failure.
func WithKnativeRetry(retry bool) KnativeOption {
	return func(knative *knativeProcessor) {
		knative.retry = retry
	}
}
