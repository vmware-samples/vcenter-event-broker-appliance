module github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router

go 1.14

require (
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/aws/aws-sdk-go v1.27.3
	github.com/cloudevents/sdk-go/v2 v2.0.0-RC2
	github.com/google/uuid v1.1.1
	github.com/jpillora/backoff v1.0.0
	github.com/onsi/ginkgo v1.12.2
	github.com/onsi/gomega v1.10.1
	github.com/openfaas-incubator/connector-sdk v0.0.0-20200902074656-7f648543d4aa
	github.com/openfaas/faas-provider v0.15.1
	github.com/pkg/errors v0.9.1
	github.com/vmware/govmomi v0.22.2
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/openfaas-incubator/connector-sdk => github.com/embano1/connector-sdk v0.0.0-20201005194225-2a76f5b4d502
