package v1alpha1

const (
	EventProcessor = "EventProcessor"
)

type ProcessorType string

const (
	ProcessorOpenFaaS    ProcessorType = "openfaas"
	ProcessorEventBridge ProcessorType = "aws_event_bridge"
)

type Processor struct {
	Type ProcessorType `yaml:"type" json:"type" jsonschema:"enum=openfaas,enum=awsEventBridge"`
	Name string        `yaml:"name" json:"name" jsonschema:"required"`
	// +optional
	OpenFaaS *ProcessorConfigOpenFaaS `yaml:"openfaas,omitempty" json:"openfaas,omitempty" jsonschema:"oneof_required=openfaas"`
	// +optional
	EventBridge *ProcessorConfigEventBridge `yaml:"awsEventBridge,omitempty" json:"awsEventBridge,omitempty" jsonschema:"oneof_required=awsEventBridge"`
}

type ProcessorConfigOpenFaaS struct {
	Address string `yaml:"address" json:"address" jsonschema:"required,description=OpenFaaS gateway address,default=http://gateway.openfaas:8080"`
	Async   bool   `yaml:"async" json:"async" jsonschema:"required,description=Use async function invocation mode,default=false"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

type ProcessorConfigEventBridge struct {
	Region   string `yaml:"region" json:"region" jsonschema:"required,default=us-west-1"`
	EventBus string `yaml:"eventBus" json:"eventBus" jsonschema:"required,default=default"`
	// TODO (@mgasch): deprecate and support 1..n rules per given eventbus
	RuleARN string `yaml:"ruleARN" json:"ruleARN" jsonschema:"required,default=arn:aws:events:us-west-1:1234567890:rule/vmware-event-router"`
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}
