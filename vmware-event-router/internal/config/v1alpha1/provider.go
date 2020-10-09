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
)

// Provider configures the event provider
type Provider struct {
	// Type sets the event provider
	// TODO: add vcd enum and vcd section once it's implemented
	Type ProviderType `yaml:"type" json:"type" jsonschema:"enum=vcenter"`
	// Name is an identifier for the configured event provider
	Name string `yaml:"name" json:"name" jsonschema:"required"`
	// VCenter configuration settings
	// +optional
	VCenter *ProviderConfigVCenter `yaml:"vcenter,omitempty" json:"vcenter,omitempty" jsonschema:"oneof_required=vcenter"`
	// VCD configuration settings
	// TODO: uncomment once implemented
	// +optional
	// VCD *ProviderConfigVCD `yaml:"vcd,omitempty" json:"vcd,omitempty" jsonschema:"oneof_required=vcd"`
}

// ProviderConfigVCenter configures the vCenter event provider
type ProviderConfigVCenter struct {
	// Address of the vCenter server (URI)
	Address string `yaml:"address" json:"address" jsonschema:"required,default=https://my-vcenter01.domain.local/sdk"`
	// InsecureSSL enables/disables TLS certificate validation
	InsecureSSL bool `yaml:"insecureSSL" json:"insecureSSL" jsonschema:"required,default=true"`
	// Checkpoint enables/disables event replay from a checkpoint file
	Checkpoint bool `yaml:"checkpoint" json:"checkpoint" jsonschema:"description=Enable checkpointing via checkpoint file for event recovery and replay purposes"`
	// CheckpointDir sets the directory for persisting checkpoints (optional)
	CheckpointDir string `yaml:"checkpointDir,omitempty" json:"checkpointDir,omitempty" jsonschema:"description=Directory where to persist checkpoints if enabled,default=./checkpoints"`
	// Auth sets the vCenter authentication credentials
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

// TODO: add fields if needed and jsonschema information, e.g. defaults, required
/*type ProviderConfigVCD struct {
	Address string `yaml:"address" json:"address" jsonschema:"required"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}*/
