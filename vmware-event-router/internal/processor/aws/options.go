package aws

import (
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
)

// Option configures the AWS processor
// TODO: change signature to return errors
type Option func(*EventBridgeProcessor)

// WithClient uses the specified EventBridge client, e.g. useful in testing
func WithClient(client eventbridgeiface.EventBridgeAPI) Option {
	return func(aws *EventBridgeProcessor) {
		aws.EventBridgeAPI = client
	}
}
