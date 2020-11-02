package processor

// KnativeOption configures the knative processor
type KnativeOption func(*KnativeProcessor)

// WithKnativeVerbose enables verbose logging for the Knative processor
func WithKnativeVerbose(verbose bool) KnativeOption {
	return func(knative *KnativeProcessor) {
		knative.verbose = verbose
	}
}
