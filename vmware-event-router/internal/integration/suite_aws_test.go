// +build integration,aws

package integration_test

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	. "github.com/onsi/gomega"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	fakeVCenterName = "https://fakevc-01:443/sdk"
)

type receiveFunc func(stats *metrics.EventStats)

func (r receiveFunc) Receive(stats *metrics.EventStats) {
	r(stats)
}

var (
	ctx context.Context
	log *zap.SugaredLogger

	awsProcessor processor.Processor
	cfg          *config.ProcessorConfigEventBridge
	receiver     receiveFunc
)

func TestAWS(t *testing.T) {
	RegisterFailHandler(Fail)
	log = zaptest.NewLogger(t).Named("[AWS_SUITE]").Sugar()
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

	cfg = &config.ProcessorConfigEventBridge{
		EventBus: awsBus,
		Region:   awsRegion,
		RuleARN:  awsRule,
		Auth: &config.AuthMethod{
			Type: config.AWSAccessKeyAuth,
			AWSAccessKeyAuth: &config.AWSAccessKeyAuthMethod{
				AccessKey: awsAccessKey,
				SecretKey: awsSecret,
			},
		},
	}

	receiver = func(stats *metrics.EventStats) {} // noOp
})

var _ = AfterSuite(func() {})
