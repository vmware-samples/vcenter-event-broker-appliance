//go:build integration && aws

package integration_test

import (
	"context"
	"os"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
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

type mockClient struct {
	eventbridgeiface.EventBridgeAPI
	sent int32 // number events sent
}

func (m *mockClient) PutEventsWithContext(ctx aws.Context, input *eventbridge.PutEventsInput, opts ...request.Option) (*eventbridge.PutEventsOutput, error) {
	atomic.AddInt32(&m.sent, 1)
	return m.EventBridgeAPI.PutEventsWithContext(ctx, input, opts...)
}

var (
	ctx context.Context
	log *zap.SugaredLogger

	awsProcessor processor.Processor
	cfg          *config.ProcessorConfigEventBridge
	receiver     receiveFunc
	ebClient     *mockClient
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

func createClient(cfg *config.ProcessorConfigEventBridge) *mockClient {
	accessKey := cfg.Auth.AWSAccessKeyAuth.AccessKey
	secretKey := cfg.Auth.AWSAccessKeyAuth.SecretKey

	awsSessionAccessKey, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"", // a token will be created when the session is used.
		),
	})

	Expect(err).ShouldNot(HaveOccurred())

	client := eventbridge.New(awsSessionAccessKey)
	Expect(client).ToNot(BeNil())

	return &mockClient{EventBridgeAPI: client}
}
