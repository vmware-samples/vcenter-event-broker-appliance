package v1alpha1

import (
	"io"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
)

const (
	// APIVersion is the API version used by this configuration
	APIVersion = "event-router.vmware.com/v1alpha1"
	// Kind sets the resource for this configuration and associated API version
	Kind = "RouterConfig"
)

// TypeMeta sets API version and kind for this configuration
type TypeMeta struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion" jsonschema:"required,enum=event-router.vmware.com/v1alpha1"`
	Kind       string `yaml:"kind" json:"kind" jsonschema:"required,enum=RouterConfig"`
}

// ObjectMeta contains addition metadata such as name and (optional) key/value pairs (labels)
type ObjectMeta struct {
	Name   string            `yaml:"name" json:"name" jsonschema:"required"`
	Labels map[string]string `yaml:"labels,omitempty" jsonschema:""`
}

// Certificates defines custom certificate types to be used instead of the
// system (OS) defaults.
type Certificates struct {
	RootCAs []string `yaml:"rootCAs,omitempty" json:"rootCAs,omitempty" jsonschema:""`
}

// RouterConfig sets configuration for the event router
type RouterConfig struct {
	TypeMeta   `yaml:",inline" jsonschema:"required"`
	ObjectMeta `yaml:"metadata" json:"metadata" jsonschema:"required"`
	// EventProvider contains configuration information for a supported event provider
	EventProvider Provider `yaml:"eventProvider" json:"eventProvider" jsonschema:"required"`
	// EventProcessor contains configuration information for a supported event processor
	EventProcessor Processor `yaml:"eventProcessor" json:"eventProcessor" jsonschema:"required"`
	// MetricsProvider contains configuration information for a supported metrics provider
	MetricsProvider MetricsProvider `yaml:"metricsProvider" json:"metricsProvider" jsonschema:"required"`
	// Certificates contains configuration information to define certificates. This
	// section is currently only used by the vCenter event provider.
	Certificates Certificates `yaml:"certificates,omitempty" json:"certificates,omitempty" jsonschema:""`
}

// Parse parses a given configuration and returns a RouterConfig
func Parse(yamlCfg io.Reader) (*RouterConfig, error) {
	var cfg RouterConfig
	dec := yaml.NewDecoder(yamlCfg, yaml.Strict())
	err := dec.Decode(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode configuration file")
	}

	return &cfg, nil
}
