package vcsim

// Option allows for customization of the vCenter simulator event provider
type Option func(*EventStream)

// WithVerbose enables verbose logging for the vCenter simulator event provider
func WithVerbose(verbose bool) Option {
	return func(vcsim *EventStream) {
		vcsim.verbose = verbose
	}
}
