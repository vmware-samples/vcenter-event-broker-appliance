package knative

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	ceclient "github.com/cloudevents/sdk-go/v2/client"
	"github.com/embano1/waitgroup"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/injection"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	maxRetries   = 3
	retryDelay   = 100 * time.Millisecond
	waitShutdown = 5 * time.Second // wait for processing to finish during shutdown
)

// Processor implements the Processor interface
type Processor struct {
	logger.Logger
	kConfig  *rest.Config
	ceClient cloudevents.Client
	sink     string
	wg       waitgroup.WaitGroup // used in graceful shutdown

	mu      sync.RWMutex
	stopped bool // indicate whether the processor has been stopped
	stats   metrics.EventStats
}

var (
	// ErrStopped is returned when a shutdown attempt is performed and the processor has already been stopped
	ErrStopped = errors.New("processor already stopped")
	// assert we implement Processor interface
	_ processor.Processor = (*Processor)(nil)
)

// NewProcessor returns a Knative processor for the given configuration
func NewProcessor(ctx context.Context, cfg *config.ProcessorConfigKnative, ms metrics.Receiver, log logger.Logger, opts ...Option) (*Processor, error) {
	kLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		proc := strings.ToUpper(string(config.ProcessorKnative))
		kLog = zapSugared.Named(fmt.Sprintf("[%s]", proc))
	}

	if cfg == nil {
		return nil, errors.New("knative configuration must be provided")
	}

	// validate the given target/destination
	if err := cfg.Destination.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "validate configuration")
	}

	var p Processor
	// apply options
	for _, opt := range opts {
		if err := opt(&p); err != nil {
			return nil, err
		}
	}

	// assume in-cluster mode if empty
	if p.kConfig == nil {
		kCfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "get Kubernetes configuration")
		}
		p.kConfig = kCfg
	}

	// start Knative informers
	ctx = logging.WithLogger(ctx, kLog.(*zap.SugaredLogger))
	ctx, startInformers := injection.EnableInjectionOrDie(ctx, p.kConfig)
	startInformers()

	// placeholder type, works with any addressable, e.g. Broker, kService
	var source corev1.Service

	uriResolver := resolver.NewURIResolver(ctx, func(name types.NamespacedName) {})
	uri, err := uriResolver.URIFromDestinationV1(ctx, *cfg.Destination, &source)
	if err != nil {
		return nil, errors.Wrap(err, "get URI from destination")
	}

	target := uri.String()
	client, err := ceClient(target, cfg.InsecureSSL, cfg.Encoding)
	if err != nil {
		return nil, err
	}

	p.Logger = kLog
	p.ceClient = client
	p.sink = target
	p.stats = metrics.EventStats{
		Provider:    string(config.ProcessorKnative),
		Type:        config.EventProcessor,
		Address:     p.sink,
		Started:     time.Now().UTC(),
		Invocations: make(map[string]*metrics.InvocationDetails),
	}

	go p.PushMetrics(ctx, ms)
	return &p, nil
}

// ceClient returns a cloud events client for the given URI and TLS insecure value
func ceClient(uri string, insecure bool, encoding string) (ceclient.Client, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: insecure, //nolint:gosec
	}
	httpTransport := &http.Transport{TLSClientConfig: &tlsConfig}

	// Create protocol and client
	transport, err := cloudevents.NewHTTP(cloudevents.WithTarget(uri), cloudevents.WithRoundTripper(httpTransport))
	if err != nil {
		return nil, errors.Wrap(err, "create cloud events http transport")
	}

	clientOpts := []ceclient.Option{ceclient.WithUUIDs()}
	switch encoding {
	case "structured":
		clientOpts = append(clientOpts, ceclient.WithForceStructured())
	case "binary":
		clientOpts = append(clientOpts, ceclient.WithForceBinary())
	default:
		return nil, fmt.Errorf("unsupported encoding type specified: %q", encoding)
	}

	client, err := cloudevents.NewClient(transport, clientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "create cloud events client")
	}

	return client, nil
}

// Process processes a cloud event. Internal retry logic handles transient
// processing errors. If an event cannot be processed an error is returned.
func (p *Processor) Process(ctx context.Context, ce cloudevents.Event) error {
	if p.isStopped() {
		return ErrStopped
	}

	// coordinate concurrent shutdown
	p.wg.Add(1)
	defer p.wg.Done()

	p.Debugw("processing event", "eventID", ce.ID(), "event", ce)
	subject := ce.Subject()
	p.mu.Lock()
	// initialize invocation stats
	// TODO: #182
	if _, ok := p.stats.Invocations[subject]; !ok {
		p.stats.Invocations[subject] = &metrics.InvocationDetails{}
	}
	p.mu.Unlock()

	// register retry options
	ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, retryDelay, maxRetries)
	p.Infow("sending event", "eventID", ce.ID(), "subject", subject)
	result := p.ceClient.Send(ctx, ce)
	p.Debugw("got response", "eventID", ce.ID(), "response", result)

	p.mu.Lock()
	defer p.mu.Unlock()

	if !cloudevents.IsACK(result) {
		p.stats.Invocations[subject].Failure()
		return processor.NewError(config.ProcessorKnative, errors.Wrapf(result, "send event %s", ce.ID()))
	}

	p.Infow("successfully sent event", "eventID", ce.ID())
	p.stats.Invocations[subject].Success()
	return nil
}

func (p *Processor) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.mu.RLock()
			ms.Receive(&p.stats)
			p.mu.RUnlock()
		}
	}
}

// Shutdown performs a clean shutdown of the Knative processor. If the processor
// has already been stopped ErrStopped is returned.
func (p *Processor) Shutdown(_ context.Context) error {
	p.Logger.Infof("attempting graceful shutdown")
	if p.isStopped() {
		return ErrStopped
	}

	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()

	p.Logger.Infof("waiting up to %v for inflight events to finish processing", waitShutdown)
	return errors.Wrap(p.wg.WaitTimeout(waitShutdown), "shutdown")
}

// Sink returns the configured destination sink where events are sent
func (p *Processor) Sink() string {
	return p.sink
}

func (p *Processor) isStopped() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stopped
}
