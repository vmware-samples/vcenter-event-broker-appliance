package v1alpha1

const (
	// EventProvider is the identifier of an event provider
	EventProvider = "EventProvider"
)

// ProviderType represents a supported event provider
type ProviderType string

const (
	// ProviderVCenter represents the vCenter event provider
	ProviderVCenter ProviderType = "vcenter"
	ProviderVCSIM   ProviderType = "vcsim"
	ProviderWebhook ProviderType = "webhook"
)

// Provider configures the event provider
type Provider struct {
	// Type sets the event provider
	Type ProviderType `yaml:"type" json:"type" jsonschema:"enum=vcenter,enum=webhook,enum=vcsim"`
	// Name is an identifier for the configured event provider
	Name string `yaml:"name" json:"name" jsonschema:"required"`
	// VCenter configuration settings
	// +optional
	VCenter *ProviderConfigVCenter `yaml:"vcenter,omitempty" json:"vcenter,omitempty" jsonschema:"oneof_required=vcenter"`
	// VCenter simulator configuration settings
	// DEPRECATED: use provider vcenter instead
	// +optional
	VCSIM *ProviderConfigVCSIM `yaml:"vcsim,omitempty" json:"vcsim,omitempty" jsonschema:"oneof_required=vcsim"`
	// Webhook configuration settings
	// +optional
	Webhook *ProviderConfigWebhook `yaml:"webhook,omitempty" json:"webhook,omitempty" jsonschema:"oneof_required=webhook"`
}

// ProviderConfigVCenter configures the vCenter event provider
type ProviderConfigVCenter struct {
	// Address of the vCenter server (URI)
	Address string `yaml:"address" json:"address" jsonschema:"required,default=https://my-vcenter01.domain.local/sdk"`
	// InsecureSSL enables/disables TLS certificate validation
	InsecureSSL bool `yaml:"insecureSSL" json:"insecureSSL" jsonschema:"required,default=false"`
	// Checkpoint enables/disables event replay from a checkpoint file
	Checkpoint bool `yaml:"checkpoint" json:"checkpoint" jsonschema:"description=Enable checkpointing via checkpoint file for event recovery and replay purposes"`
	// CheckpointDir sets the directory for persisting checkpoints (optional)
	CheckpointDir string `yaml:"checkpointDir,omitempty" json:"checkpointDir,omitempty" jsonschema:"description=Directory where to persist checkpoints if enabled,default=./checkpoints"`
	// Auth sets the vCenter authentication credentials
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

// ProviderConfigVCSIM configures the vCenter simulator event provider
type ProviderConfigVCSIM struct {
	// Address of the vCenter simulator server (URI)
	Address string `yaml:"address" json:"address" jsonschema:"required,default=https://my-vcenter01.domain.local/sdk"`
	// InsecureSSL enables/disables TLS certificate validation
	InsecureSSL bool `yaml:"insecureSSL" json:"insecureSSL" jsonschema:"required,default=false"`
	// Auth sets the vCenter simulator authentication credentials
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

// ProviderConfigWebhook configures the webhook event provider
type ProviderConfigWebhook struct {
	// BindAddress is the address where the webhook http server will listen for
	// connections
	BindAddress string `yaml:"bindAddress" json:"bindAddress" jsonschema:"required,default=0.0.0.0:8080"`
	// Path is the relative URL path to accept incoming webhook CloudEvents
	Path string `yaml:"path" json:"path" jsonschema:"required,default=/webhook"`

	// Auth sets the webhook authentication credentials
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`

	// 	TODO: concurrency (goroutines), inbound/outbound rate limit
}
