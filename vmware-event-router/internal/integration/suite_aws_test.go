// +build integration,aws

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

const (
	fakeVCenterName = "https://fakevc-01:443/sdk"
)

// implement metrics interface
type fakeReceiver struct {
}

func (f *fakeReceiver) Receive(stats metrics.EventStats) {
}

var (
	ctx          context.Context
	awsProcessor processor.Processor
	receiver     *fakeReceiver
)

func TestAWS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS EventBridge Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY")
	awsSecret := os.Getenv("AWS_SECRET_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	awsBus := os.Getenv("AWS_EVENT_BUS")
	awsRule := os.Getenv("AWS_RULE_ARN")

	Expect(awsAccessKey).ToNot(BeEmpty(), "env var AWS_ACCESS_KEY to authenticate against AWS EventBridge must be set")
	Expect(awsSecret).ToNot(BeEmpty(), "env var AWS_SECRET_KEY to authenticate against AWS EventBridge must be set")
	Expect(awsRegion).ToNot(BeEmpty(), "env var AWS_REGION for AWS EventBridge must be set")
	Expect(awsBus).ToNot(BeEmpty(), "env var AWS_EVENT_BUS for AWS EventBridge must be set")
	Expect(awsRule).ToNot(BeEmpty(), "env var AWS_RULE_ARN for AWS EventBridge must be set")

	cfg := connection.Config{
		Type: "processor",
		Options: map[string]string{
			"aws_region":                awsRegion,
			"aws_eventbridge_event_bus": awsBus,
			"aws_eventbridge_rule_arn":  awsRule,
		},
		Provider: "aws_event_bridge",
		Auth: connection.Authentication{
			Method: "access_key",
			Secret: map[string]string{
				"aws_access_key_id":     awsAccessKey,
				"aws_secret_access_key": awsSecret,
			},
		},
	}

	receiver = &fakeReceiver{}
	p, err := processor.NewAWSEventBridgeProcessor(ctx, cfg, receiver, processor.WithAWSVerbose(true))
	Expect(err).NotTo(HaveOccurred())
	awsProcessor = p
})

var _ = AfterSuite(func() {})
