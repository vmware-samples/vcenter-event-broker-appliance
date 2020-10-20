module github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router

go 1.14

require (
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/aws/aws-sdk-go v1.33.12
	github.com/cloudevents/sdk-go/v2 v2.0.0-RC2
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/uuid v1.1.1
	github.com/jpillora/backoff v1.0.0
	github.com/onsi/ginkgo v1.12.2
	github.com/onsi/gomega v1.10.1
	github.com/openfaas-incubator/connector-sdk v0.0.0-20200902074656-7f648543d4aa
	github.com/openfaas/faas-provider v0.15.1
	github.com/pkg/errors v0.9.1
	github.com/vmware/govmomi v0.22.2
	go.opencensus.io v0.22.3 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200728102440-3e129f6d46b1 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200725200936-102e7d357031 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
)

replace github.com/openfaas-incubator/connector-sdk => github.com/embano1/connector-sdk v0.0.0-20201013145543-3487b96b0f91
