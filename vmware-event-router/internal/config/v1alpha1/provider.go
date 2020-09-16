package v1alpha1

const (
	EventProvider = "EventProvider"
)

type ProviderType string

const (
	ProviderVCenter ProviderType = "vcenter"
	ProviderVCD     ProviderType = "vcd"
)

type Provider struct {
	// TODO: add vcd enum and vcd section once it's implemented
	Type ProviderType `yaml:"type" json:"type" jsonschema:"enum=vcenter"`
	Name string       `yaml:"name" json:"name" jsonschema:"required"`
	// +optional
	VCenter *ProviderConfigVCenter `yaml:"vcenter,omitempty" json:"vcenter,omitempty" jsonschema:"oneof_required=vcenter"`
	// TODO: uncomment once implemented
	// +optional
	// VCD *ProviderConfigVCD `yaml:"vcd,omitempty" json:"vcd,omitempty" jsonschema:"oneof_required=vcd"`
}

type ProviderConfigVCenter struct {
	Address     string `yaml:"address" json:"address" jsonschema:"required,default=https://my-vcenter01.domain.local/sdk"`
	InsecureSSL bool   `yaml:"insecureSSL" json:"insecureSSL" jsonschema:"required,default=true"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

// TODO: add fields if needed and jsonschema information, e.g. defaults, required
type ProviderConfigVCD struct {
	Address string `yaml:"address" json:"address" jsonschema:"required"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}
