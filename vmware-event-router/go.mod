module github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router

go 1.14

require (
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/aws/aws-sdk-go v1.33.12
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/embano1/waitgroup v0.0.0-20201120223302-1d5df9b49112
	github.com/goccy/go-yaml v1.8.4
	github.com/google/uuid v1.1.2
	github.com/jpillora/backoff v1.0.0
	github.com/onsi/ginkgo v1.12.2
	github.com/onsi/gomega v1.10.1
	github.com/openfaas-incubator/connector-sdk v0.0.0-20200902074656-7f648543d4aa
	github.com/openfaas/faas-provider v0.15.1
	github.com/pkg/errors v0.9.1
	github.com/vmware/govmomi v0.24.1-0.20210210035757-ed60338583b0
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	knative.dev/pkg v0.0.0-20201109175709-2c9320ae0640
)

replace github.com/openfaas-incubator/connector-sdk => github.com/embano1/connector-sdk v0.0.0-20201209211641-e6a3409ab348
