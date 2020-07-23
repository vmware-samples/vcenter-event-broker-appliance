// +build integration,openfaas

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"

	ofsdk "github.com/openfaas-incubator/connector-sdk/types"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

const (
	fakeVCenterName = "https://fakevc-01:443/sdk"
	success         = true
	fail            = false
)

// implement metrics and invoker response handler
type fakeReceiver struct {
	sync.RWMutex
	invocations int
	responseMap map[bool]int
}

func (f *fakeReceiver) Receive(stats metrics.EventStats) {
	f.Lock()
	defer f.Unlock()
	f.invocations++
}

func (f *fakeReceiver) Response(res ofsdk.InvokerResponse) {
	f.Lock()
	defer f.Unlock()
	if res.Error != nil || res.Status != http.StatusOK {
		fmt.Fprintf(GinkgoWriter, "function %s for topic %s returned status %d with error: %v", res.Function, res.Topic, res.Status, res.Error)
		f.responseMap[fail]++
		return
	}
	fmt.Fprintf(GinkgoWriter, "successfully invoked function %s for topic %s", res.Function, res.Topic)
	f.responseMap[success]++
}

var (
	ctx         context.Context
	ofProcessor processor.Processor
	receiver    *fakeReceiver
)

func TestOpenfaas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Openfaas Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	receiver = &fakeReceiver{}

	// we assume basic_auth and fail if OpenFaaS gateway password is empty
	ofPass := os.Getenv("OF_PASSWORD")
	Expect(ofPass).ToNot(BeEmpty(), "env var OF_PASSWORD for basic_auth against OpenFaaS gateway must be set")

	cfg := connection.Config{
		Type:    "processor",
		Address: "http://localhost:8080",
		Options: map[string]string{
			"async": "false",
		},
		Provider: "openfaas",
		Auth: connection.Authentication{
			Method: "basic_auth",
			Secret: map[string]string{
				"username": "admin",
				"password": ofPass,
			},
		},
	}
	op, err := processor.NewOpenFaaSProcessor(ctx,
		cfg,
		receiver,
		processor.WithOpenFaaSRebuildInterval(100*time.Millisecond),
		processor.WithOpenFaaSResponseHandler(receiver),
	)
	Expect(err).ShouldNot(HaveOccurred())
	ofProcessor = op
})

var _ = AfterSuite(func() {})
