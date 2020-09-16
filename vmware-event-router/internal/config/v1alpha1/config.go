package v1alpha1

import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	APIVersion = "event-router.vmware.com/v1alpha1"
	Kind       = "RouterConfig"
)

type TypeMeta struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion" jsonschema:"required,enum=event-router.vmware.com/v1alpha1"`
	Kind       string `yaml:"kind" json:"kind" jsonschema:"required,enum=RouterConfig"`
}

type ObjectMeta struct {
	Name   string            `yaml:"name" json:"name" jsonschema:"required"`
	Labels map[string]string `yaml:"labels,omitempty" jsonschema:""`
}

type RouterConfig struct {
	TypeMeta        `yaml:",inline" jsonschema:"required"`
	ObjectMeta      `yaml:"metadata" json:"metadata" jsonschema:"required"`
	EventProvider   Provider        `yaml:"eventProvider" json:"eventProvider" jsonschema:"required"`
	EventProcessor  Processor       `yaml:"eventProcessor" json:"eventProcessor" jsonschema:"required"`
	MetricsProvider MetricsProvider `yaml:"metricsProvider" json:"metricsProvider" jsonschema:"required"`
}

func Parse(yamlCfg io.Reader) (*RouterConfig, error) {
	var cfg RouterConfig
	dec := yaml.NewDecoder(yamlCfg)
	err := dec.Decode(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode configuration file")
	}

	return &cfg, nil
}
