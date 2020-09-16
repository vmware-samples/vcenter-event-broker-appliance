package v1alpha1

type AuthMethodType string

const (
	BasicAuth        AuthMethodType = "basic_auth"
	AWSAccessKeyAuth AuthMethodType = "aws_access_key"
)

type AuthMethod struct {
	Type AuthMethodType `yaml:"type" json:"type" jsonschema:"enum=basic_auth,enum=aws_access_key,default=basic_auth,description=The authentication method to use"`
	// +optional
	BasicAuth *BasicAuthMethod `yaml:"basicAuth,omitempty" json:"basicAuth,omitempty" jsonschema:"oneof_required=basicAuth,description=Basic authentication with username and password"`
	// +optional
	AWSAccessKeyAuth *AWSAccessKeyAuthMethod `yaml:"awsAccessKeyAuth,omitempty" json:"awsAccessKeyAuth,omitempty" jsonschema:"oneof_required=awsAccessKeyAuth,description=AWS authentication with access and secret key"`
}

type BasicAuthMethod struct {
	Username string `yaml:"username" json:"username" jsonschema:"required"`
	Password string `yaml:"password" json:"password" jsonschema:"required"`
}

type AWSAccessKeyAuthMethod struct {
	AccessKey string `yaml:"accessKey" json:"accessKey" jsonschema:"required"`
	SecretKey string `yaml:"secretKey" json:"secretKey" jsonschema:"required"`
}
