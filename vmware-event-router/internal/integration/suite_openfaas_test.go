//go:build integration && openfaas

package integration_test

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"

	ofsdk "github.com/openfaas-incubator/connector-sdk/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"

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

func (f *fakeReceiver) Receive(_ *metrics.EventStats) {
	f.Lock()
	defer f.Unlock()
	f.invocations++
}

func (f *fakeReceiver) Response(res ofsdk.InvokerResponse) {
	f.Lock()
	defer f.Unlock()

	// when finished send response to unblock and return from Process() in the
	// processor
	defer func() {
		resCh <- res
	}()

	if res.Error != nil || res.Status != http.StatusOK {
		f.responseMap[fail]++
		return
	}

	f.responseMap[success]++
}

var (
	ctx = context.Background()
	log *zap.SugaredLogger

	ofProcessor processor.Processor
	cfg         *config.ProcessorConfigOpenFaaS
	receiver    *fakeReceiver
	resCh       = make(chan ofsdk.InvokerResponse)
)

func TestOpenfaas(t *testing.T) {
	RegisterFailHandler(Fail)
	log = zaptest.NewLogger(t).Named("[OPENFAAS_SUITE]").Sugar()
	RunSpecs(t, "OpenFaaS Suite")
}

var _ = BeforeSuite(func() {
	// we assume basic_auth and fail if OpenFaaS gateway password is empty
	ofPass := os.Getenv("OF_PASSWORD")
	Expect(ofPass).ToNot(BeEmpty(), "env var OF_PASSWORD for basic_auth against OpenFaaS gateway must be set")

	cfg = &config.ProcessorConfigOpenFaaS{
		Address: "http://localhost:8080",
		Async:   false,
		Auth: &config.AuthMethod{
			Type: config.BasicAuth,
			BasicAuth: &config.BasicAuthMethod{
				Username: "admin",
				Password: ofPass,
			},
		},
	}
})

var _ = AfterSuite(func() {
	close(resCh)
})
