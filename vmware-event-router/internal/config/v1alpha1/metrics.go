package v1alpha1

type MetricsProviderType string

const (
	MetricsProviderDefault MetricsProviderType = "default"
)

type MetricsProvider struct {
	Type MetricsProviderType `yaml:"type" json:"type" jsonschema:"enum=default"`
	Name string              `yaml:"name" json:"name" jsonschema:"required"`
	// +optional
	Default *MetricsProviderConfigDefault `yaml:"default,omitempty" json:"default,omitempty" jsonschema:"oneof_required=default"`
}

type MetricsProviderConfigDefault struct {
	BindAddress string `yaml:"bindAddress" json:"bindAddress" jsonschema:"required"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`
}
