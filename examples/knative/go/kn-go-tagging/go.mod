module github.com/vmware-samples/vcenter-event-broker-appliance/examples/knative/go/kn-go-tagging

go 1.16

require (
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/vmware/govmomi v0.27.2
	go.uber.org/zap v1.19.1
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.22.2
	knative.dev/pkg v0.0.0-20211026134021-7049a59d8e37
)
