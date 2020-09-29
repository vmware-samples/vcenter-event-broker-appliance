package vcenter

// Option allows for customization of the vCenter event provider
type Option func(*EventStream)

// WithVerbose enables verbose logging for the AWS processor
func WithVerbose(verbose bool) Option {
	return func(vc *EventStream) {
		vc.verbose = verbose
	}
}
