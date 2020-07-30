// +build integration,openfaas

package integration_test

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// how long to wait for gatway response/topic map sync
	waitDelay = 1 * time.Second
)

var _ = Describe("OpenFaaS Processor", func() {
	BeforeEach(func() {
		// give topic map builder in the controller time to sync
		time.Sleep(waitDelay)
	})

	BeforeEach(func() {
		// reset map
		receiver.responseMap = make(map[bool]int)
	})

	Describe("receiving an event", func() {
		var (
			baseEvent    types.BaseEvent
			ce           *cloudevents.Event
			err          error
			successCount int
			failCount    int
		)

		Context("when one function in OpenFaaS is subscribed to that event type (VmPoweredOnEvent)", func() {
			// create VMPoweredOnEvent and marshal to CloudEvent
			BeforeEach(func() {
				baseEvent = newVMPoweredOnEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			// process and give response time to get back from OpenFaaS gateway
			BeforeEach(func() {
				err = ofProcessor.Process(*ce)
				time.Sleep(waitDelay)
			})

			// avoid races
			BeforeEach(func() {
				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should receive a successful invokation response", func() {
				Expect(successCount).To(Equal(1))
			})

			It("should not receive a failed invokation response", func() {
				Expect(failCount).To(Equal(0))
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when a subscribed function in OpenFaaS returns an error (ClusterCreatedEvent)", func() {
			// create ClusterCreatedEvent and marshal to CloudEvent
			BeforeEach(func() {
				baseEvent = newClusterCreatedEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			// process and give response time to get back from OpenFaaS gateway
			BeforeEach(func() {
				err = ofProcessor.Process(*ce)
				time.Sleep(waitDelay)
			})

			// avoid races
			BeforeEach(func() {
				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should not receive a successful invokation response", func() {
				Expect(successCount).To(Equal(0))
			})

			It("should receive a failed invokation response", func() {
				Expect(failCount).To(Equal(1))
			})

			// OpenFaaS processor does not return error directly, only in case
			// of JSON marhalling issues which we don't expect for this test case
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when no function in OpenFaaS is subscribed to that event type (LicenseEvent)", func() {
			// create LicenseEvent and marshal to CloudEvent
			BeforeEach(func() {
				baseEvent = newLicenseEvent()
				ce, err = events.NewCloudEvent(baseEvent, fakeVCenterName)
				Expect(err).ShouldNot(HaveOccurred())
			})

			// process and give response time to get back from OpenFaaS gateway
			// (note: we don't expect response, just making sure nothing gets
			// through)
			BeforeEach(func() {
				err = ofProcessor.Process(*ce)
				time.Sleep(waitDelay)
			})

			// avoid races
			BeforeEach(func() {
				receiver.RLock()
				defer receiver.RUnlock()
				successCount = receiver.responseMap[success]
				failCount = receiver.responseMap[fail]
			})

			It("should not receive a successful invokation response", func() {
				Expect(successCount).To(Equal(0))
			})

			It("should receive a failed invokation response", func() {
				Expect(failCount).To(Equal(0))
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
