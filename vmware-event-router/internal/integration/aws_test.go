// +build integration,aws

package integration_test

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor/aws"
)

var _ = Describe("AWS Processor", func() {
	BeforeEach(func() {
		p, err := aws.NewEventBridgeProcessor(ctx, cfg, receiver, log)
		Expect(err).NotTo(HaveOccurred())
		awsProcessor = p

		// give pattern match map time to sync
		time.Sleep(time.Second)
	})

	Describe("receiving an event", func() {
		var (
			baseEvent types.BaseEvent
			ce        *cloudevents.Event
			err       error
		)

		Context("when the EventBridge rule pattern matches the given event type (VmPoweredOnEvent)", func() {
			// create VMPoweredOnEvent and marshal to CloudEvent
			BeforeEach(func() {
				baseEvent = newVMPoweredOnEvent()
				ce, err = events.NewFromVSphere(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			// process
			BeforeEach(func() {
				err = awsProcessor.Process(ctx, *ce)
			})

			It("should not error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the EventBridge rule pattern does not match the given event type (LicenseEvent)", func() {
			// create LicenseEvent and marshal to CloudEvent
			BeforeEach(func() {
				baseEvent = newLicenseEvent()
				ce, err = events.NewFromVSphere(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			// process
			BeforeEach(func() {
				err = awsProcessor.Process(ctx, *ce)
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
