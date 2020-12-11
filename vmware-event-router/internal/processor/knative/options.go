package knative

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

// Option configures the Knative processor
type Option func(*Processor) error

// WithRestConfig provides a custom Kubernetes REST configuration, e.g. for
// out-of-cluster configurations
func WithRestConfig(cfg *rest.Config) Option {
	return func(o *Processor) error {
		if cfg == nil {
			return errors.New("no config provided")
		}
		o.kConfig = cfg
		return nil
	}
}
