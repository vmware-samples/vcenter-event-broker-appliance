package v1alpha1

// AuthMethodType represents a supported authentication method
type AuthMethodType string

const (
	// BasicAuth represents the basic authentication method
	BasicAuth AuthMethodType = "basic_auth"
	// AWSAccessKeyAuth represents the AWS IAM authentication method using access key and secret key
	AWSAccessKeyAuth AuthMethodType = "aws_access_key"
)

// AuthMethod configures authentication data
type AuthMethod struct {
	// Type sets the authentication method
	Type AuthMethodType `yaml:"type" json:"type" jsonschema:"enum=basic_auth,enum=aws_access_key,default=basic_auth,description=The authentication method to use"`
	// +optional
	BasicAuth *BasicAuthMethod `yaml:"basicAuth,omitempty" json:"basicAuth,omitempty" jsonschema:"oneof_required=basicAuth,description=Basic authentication with username and password"`
	// +optional
	AWSAccessKeyAuth *AWSAccessKeyAuthMethod `yaml:"awsAccessKeyAuth,omitempty" json:"awsAccessKeyAuth,omitempty" jsonschema:"oneof_required=awsAccessKeyAuth,description=AWS authentication with access and secret key"`
}

// BasicAuthMethod configures authentication data for BasicAuth
type BasicAuthMethod struct {
	Username string `yaml:"username" json:"username" jsonschema:"required"`
	Password string `yaml:"password" json:"password" jsonschema:"required"`
}

// AWSAccessKeyAuthMethod configures authentication data for AWSAccessKeyAuth
type AWSAccessKeyAuthMethod struct {
	AccessKey string `yaml:"accessKey" json:"accessKey" jsonschema:"required"`
	SecretKey string `yaml:"secretKey" json:"secretKey" jsonschema:"required"`
}
