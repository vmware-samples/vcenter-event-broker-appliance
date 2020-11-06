// +build integration,openfaas

package integration_test

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor/openfaas"
)

var _ = Describe("OpenFaaS Processor", func() {
	BeforeEach(func() {
		receiver = &fakeReceiver{
			responseMap: make(map[bool]int),
		}

		log.Info("creating new processor")
		op, err := openfaas.NewProcessor(ctx,
			cfg,
			receiver,
			log,
			openfaas.WithRebuildInterval(time.Millisecond),
			openfaas.WithResponseHandler(receiver),
			openfaas.WithResponseChan(resCh),
		)
		Expect(err).ShouldNot(HaveOccurred())
		ofProcessor = op

		// give topic map builder in the controller time to sync
		time.Sleep(time.Second)
	})

	Describe("receiving an event", func() {
		var (
			baseEvent    types.BaseEvent
			ce           *cloudevents.Event
			err          error
			successCount int
			failCount    int
		)

		// assumes only of-echo is subscribed to this event
		Context("when one function in OpenFaaS is subscribed to that event type (VmPoweredOnEvent)", func() {
			BeforeEach(func() {
				baseEvent = newVMPoweredOnEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			BeforeEach(func() {
				err = ofProcessor.Process(ctx, *ce)
				Expect(err).NotTo(HaveOccurred())

				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should receive a successful invocation response", func() {
				Expect(successCount).To(Equal(1))
			})

			It("should not receive a failed invocation response", func() {
				Expect(failCount).To(Equal(0))
			})
		})

		// assumes only of-fail is subscribed to this event
		Context("when a subscribed function in OpenFaaS returns an error (ClusterCreatedEvent)", func() {
			BeforeEach(func() {
				baseEvent = newClusterCreatedEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			BeforeEach(func() {
				err = ofProcessor.Process(ctx, *ce)
				// OpenFaaS processor does not return error directly, only in case
				// of JSON marshaling issues which we don't expect for this test case
				Expect(err).NotTo(HaveOccurred())

				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should not receive a successful invocation response", func() {
				Expect(successCount).To(Equal(0))
			})

			It("should receive a failed invocation response", func() {
				Expect(failCount).To(Equal(1))
			})
		})

		Context("when no function in OpenFaaS is subscribed to that event type (LicenseEvent)", func() {
			BeforeEach(func() {
				baseEvent = newLicenseEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			BeforeEach(func() {
				err = ofProcessor.Process(ctx, *ce)
				Expect(err).NotTo(HaveOccurred())

				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should not receive a successful invocation response", func() {
				Expect(successCount).To(Equal(0))
			})

			It("should not receive a failed invocation response", func() {
				Expect(failCount).To(Equal(0))
			})
		})
	})
})
