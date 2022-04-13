package v1alpha1

// AuthMethodType represents a supported authentication method
type AuthMethodType string

const (
	// BasicAuth represents the basic authentication method
	BasicAuth AuthMethodType = "basic_auth"
	// AWSAccessKeyAuth represents the AWS IAM authentication method using access key and secret key
	AWSAccessKeyAuth AuthMethodType = "aws_access_key"
	// AWSIAMRoleAuth represents the AWS IAM authentication method using temporary credentials provided
	// by Security Token Service (STS). Intended for use as IAM role with a Kubernetes service account
	// for use case of running under the Amazon EKS.
	AWSIAMRoleAuth AuthMethodType = "aws_iam_role"
	// 	ActiveDirectory represents the MS Active Directory domain/user/password scheme
	ActiveDirectory AuthMethodType = "active_directory"
)

// AuthMethod configures authentication data
type AuthMethod struct {
	// Type sets the authentication method
	Type AuthMethodType `yaml:"type" json:"type" jsonschema:"enum=basic_auth,enum=aws_access_key,enum=aws_iam_role,enum=active_directory,default=basic_auth,description=The authentication method to use"`
	// +optional
	BasicAuth *BasicAuthMethod `yaml:"basicAuth,omitempty" json:"basicAuth,omitempty" jsonschema:"oneof_required=basicAuth,description=Basic authentication with username and password"`
	// +optional
	AWSAccessKeyAuth *AWSAccessKeyAuthMethod `yaml:"awsAccessKeyAuth,omitempty" json:"awsAccessKeyAuth,omitempty" jsonschema:"oneof_required=awsAccessKeyAuth,description=AWS authentication with access and secret key"`
	// +optional
	ActiveDirectoryAuth *ActiveDirectoryAuthMethod `yaml:"activeDirectoryAuth,omitempty" json:"activeDirectoryAuth,omitempty" jsonschema:"oneof_required=activeDirectoryAuth,description=Active Directory authentication with domain, username and password"`
}

// BasicAuthMethod configures authentication data for basic_auth
type BasicAuthMethod struct {
	Username string `yaml:"username" json:"username" jsonschema:"required"`
	Password string `yaml:"password" json:"password" jsonschema:"required"`
}

// AWSAccessKeyAuthMethod configures authentication data for aws_access_key
type AWSAccessKeyAuthMethod struct {
	AccessKey string `yaml:"accessKey" json:"accessKey" jsonschema:"required"`
	SecretKey string `yaml:"secretKey" json:"secretKey" jsonschema:"required"`
}

// ActiveDirectoryAuthMethod configures authentication data for
// active_directory.
type ActiveDirectoryAuthMethod struct {
	Domain   string `yaml:"domain" json:"domain" jsonschema:"required"`
	Username string `yaml:"username" json:"username" jsonschema:"required"`
	Password string `yaml:"password" json:"password" jsonschema:"required"`
}
