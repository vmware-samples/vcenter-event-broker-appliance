package v1alpha1

// MetricsProviderType represents a supported metrics provider
type MetricsProviderType string

const (
	// MetricsProviderDefault is the the default metrics provider
	MetricsProviderDefault MetricsProviderType = "default"
)

// MetricsProvider configures the metrics provider
type MetricsProvider struct {
	// Type sets the metrics provider
	Type MetricsProviderType `yaml:"type" json:"type" jsonschema:"enum=default"`
	// Name is an identifier for the configured metrics provider
	Name string `yaml:"name" json:"name" jsonschema:"required"`
	// +optional
	Default *MetricsProviderConfigDefault `yaml:"default,omitempty" json:"default,omitempty" jsonschema:"oneof_required=default"`
}

// MetricsProviderConfigDefault configures the default metrics provider
type MetricsProviderConfigDefault struct {
	// BindAddress is the address where the default metrics provider http endpoint will listen for connections
	BindAddress string `yaml:"bindAddress" json:"bindAddress" jsonschema:"required"`
	// Auth when specified requires authentication for the http endpoint of the metrics provider
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`
}
