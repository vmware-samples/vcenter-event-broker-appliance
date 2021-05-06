package vcenter

// Option allows for customization of the vCenter event provider
// TODO: change signature to return errors
type Option func(*EventStream)

// WithRootCAs configures the TLS transport cert pool to use the specified root
// CA PEM-files instead of the host OS system default
func WithRootCAs(pemFiles []string) Option {
	return func(stream *EventStream) {
		stream.rootCAs = pemFiles
	}
}
