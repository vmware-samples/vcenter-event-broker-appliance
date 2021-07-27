package v1alpha1

import duckv1 "knative.dev/pkg/apis/duck/v1"

const (
	// EventProcessor is the identifier of an event processor
	EventProcessor = "EventProcessor"
)

// ProcessorType represents a supported event processor
type ProcessorType string

const (
	// ProcessorOpenFaaS represents the OpenFaaS event processor
	ProcessorOpenFaaS ProcessorType = "openfaas"
	// ProcessorEventBridge represents the AWS Event Bridge event processor
	ProcessorEventBridge ProcessorType = "aws_event_bridge"
	// ProcessorKnative represents the Knative event processor
	ProcessorKnative ProcessorType = "knative"
)

// Processor configures the event processor
type Processor struct {
	// Type sets the event processor
	Type ProcessorType `yaml:"type" json:"type" jsonschema:"enum=openfaas,enum=aws_event_bridge,enum=knative,required"`
	// Name is an identifier for the configured event processor
	Name string `yaml:"name" json:"name" jsonschema:"required"`
	// OpenFaaS configuration settings
	// +optional
	OpenFaaS *ProcessorConfigOpenFaaS `yaml:"openfaas,omitempty" json:"openfaas,omitempty" jsonschema:"oneof_required=openfaas"`
	// EventBridge configuration settings
	// +optional
	EventBridge *ProcessorConfigEventBridge `yaml:"awsEventBridge,omitempty" json:"awsEventBridge,omitempty" jsonschema:"oneof_required=awsEventBridge"`
	// Knative configuration settings
	// +optional
	Knative *ProcessorConfigKnative `yaml:"knative,omitempty" json:"knative,omitempty" jsonschema:"oneof_required=knative"`
}

// ProcessorConfigOpenFaaS configures the OpenFaaS event processor
type ProcessorConfigOpenFaaS struct {
	// Address is the connection address to the OpenFaaS gateway
	Address string `yaml:"address" json:"address" jsonschema:"required,description=OpenFaaS gateway address,default=http://gateway.openfaas:8080"`
	// Async enables/disables async function invocation mode
	Async bool `yaml:"async" json:"async" jsonschema:"required,description=Use async function invocation mode,default=false"`
	// Auth sets the OpenFaaS authentication credentials (optional). Only basic_auth
	// is supported. +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`
}

// ProcessorConfigEventBridge configures the AWS Event Bridge event processor
type ProcessorConfigEventBridge struct {
	// Region is the AWS Region of this AWS Event Bridge instance
	Region string `yaml:"region" json:"region" jsonschema:"required,default=us-west-1"`
	// EventBus is the name of the event bus (or "default" for the default event bus)
	EventBus string `yaml:"eventBus" json:"eventBus" jsonschema:"required,default=default"`
	// RuleARN is the ARN of the rule to use for configuring pattern matching and event forwarding
	// TODO (@mgasch): deprecate and support 1..n rules per given eventbus
	RuleARN string `yaml:"ruleARN" json:"ruleARN" jsonschema:"required,default=arn:aws:events:us-west-1:1234567890:rule/vmware-event-router"`
	// Auth sets the AWS authentication credentials. Only aws_access_key is
	// supported.
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"oneof_required=auth,description=Authentication configuration for this section"`
}

// ProcessorConfigKnative configures the Knative event processor
type ProcessorConfigKnative struct {
	// Destination defines the sink where events are sent to
	Destination *duckv1.Destination `yaml:"destination,omitempty" json:"destination,omitempty" jsonschema:"oneof_required=destination,description=Destination sink where to send events"`
	// InsecureSSL enables/disables TLS certificate validation
	InsecureSSL bool `yaml:"insecureSSL" json:"insecureSSL" jsonschema:"required,default=false"`
	// Encoding set the cloud event encoding type
	Encoding string `yaml:"encoding" json:"encoding" jsonschema:"enum=binary,enum=structured,required,default=structured"`
	// Auth sets the Knative authentication credentials for the specified destination
	// +optional
	Auth *AuthMethod `yaml:"auth,omitempty" json:"auth,omitempty" jsonschema:"description=Authentication configuration for this section"`
}
