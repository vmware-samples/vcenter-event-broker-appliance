package stream

import "log"

// VCenterOption configures the vCenter stream provider
type VCenterOption func(*vCenterStream)

// WithVCenterVerbose enables verbose logging for the AWS processor
func WithVCenterVerbose(verbose bool) VCenterOption {
	return func(vc *vCenterStream) {
		vc.verbose = verbose
	}
}

// WithVCenterLogger sets an alternative logger for the vCenter stream provider
func WithVCenterLogger(logger *log.Logger) VCenterOption {
	return func(vc *vCenterStream) {
		vc.Logger = logger
	}
}