package horizon

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
)

const (
	defaultPollInterval = time.Second
	eventTypeScheme     = "%s/horizon.%s.v0" // router prefix + normalized event type
)

var (
	defaultBackoff = backoff.Backoff{
		Factor: 2,
		Jitter: false,
		Min:    time.Second,
		Max:    5 * time.Second,
	}
)

// EventStream handles the connection to the Horizon events API
type EventStream struct {
	client        Client
	clock         clock.Clock
	pollInterval  time.Duration
	backoffConfig *backoff.Backoff
	logger.Logger

	sync.RWMutex
	stats metrics.EventStats
}

// NewEventStream returns a Horizon event stream manager for the given
// configuration and metrics server
func NewEventStream(ctx context.Context, cfg *config.ProviderConfigHorizon, ms metrics.Receiver, log logger.Logger, opts ...Option) (*EventStream, error) {
	if cfg == nil {
		return nil, errors.New("horizon configuration must be provided")
	}

	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "address invalid: %q", cfg.Address)
	}

	// catches parsing errors which url.Parse won't err out
	if u.Host == "" {
		return nil, fmt.Errorf("address invalid: %q", cfg.Address)
	}

	auth := cfg.Auth
	if auth == nil || auth.ActiveDirectoryAuth == nil {
		return nil, fmt.Errorf("invalid %s credentials: domain, username and password must be set", config.ActiveDirectory)
	}

	if auth.Type != config.ActiveDirectory {
		return nil, fmt.Errorf("invalid authentication type specified: %q", cfg.Auth.Type)
	}

	authDetails := auth.ActiveDirectoryAuth
	// verify all fields (incl. domain) are set as UPN login is not supported
	emptyCredentials := func() bool {
		if authDetails.Domain == "" || authDetails.Username == "" || authDetails.Password == "" {
			return true
		}
		return false
	}

	if emptyCredentials() {
		return nil, fmt.Errorf("invalid %s credentials: domain, username and password must be set", config.ActiveDirectory)
	}

	stream := EventStream{
		Logger:       log,
		clock:        clock.New(),
		pollInterval: defaultPollInterval,
	}

	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		prov := strings.ToUpper(string(config.ProviderHorizon))
		stream.Logger = zapSugared.Named(fmt.Sprintf("[%s]", prov))
		ctx = logging.WithLogger(ctx, stream.Logger.(*zap.SugaredLogger))
	}

	creds := AuthLoginRequest{
		Domain:   authDetails.Domain,
		Username: authDetails.Username,
		Password: authDetails.Password,
	}

	client, err := newHorizonClient(ctx, u.String(), creds, cfg.InsecureSSL, stream.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "create horizon API client")
	}
	stream.client = client

	stream.stats = metrics.EventStats{
		Provider:    string(config.ProviderHorizon),
		Type:        config.EventProvider,
		Address:     u.String(),
		Started:     time.Now().UTC(),
		EventsTotal: new(int),
		EventsErr:   new(int),
		EventsSec:   new(float64),
	}

	// apply options (overwrite defaults)
	for _, opt := range opts {
		opt(&stream)
	}

	go stream.PushMetrics(ctx, ms)

	return &stream, nil
}

// PushMetrics periodically pushes metrics to the metrics server
func (es *EventStream) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := es.clock.Ticker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			es.Lock()
			eventsSec := math.Round((float64(*es.stats.EventsTotal)/time.Since(es.stats.Started).Seconds())*100) / 100 // 0.2f syntax
			es.stats.EventsSec = &eventsSec
			ms.Receive(&es.stats)
			es.Unlock()
		}
	}
}

// Stream starts the event stream and polls the Horizon event API until the
// specified context is cancelled
func (es *EventStream) Stream(ctx context.Context, p processor.Processor) error {
	var (
		lastEvent *AuditEventSummary
		since     Timestamp
	)

	if es.backoffConfig == nil {
		es.backoffConfig = &defaultBackoff
	}

	ticker := es.clock.Ticker(es.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			es.Logger.Infof("stopping event stream")
			return ctx.Err()

		case <-ticker.C:
			if lastEvent != nil {
				since = Timestamp(lastEvent.Time)
			}

			if since == 0 {
				es.Debug("retrieving initial set of events")
			} else {
				es.Debugw("retrieving events with time range filter", "sinceUnixMilli", since, "sinceConverted", time.Unix(int64(since/1000), 0).String())
			}

			ev, err := es.client.GetEvents(ctx, since)
			if err != nil {
				return errors.Wrap(err, "get events")
			}

			// check if returned event is same as last event
			if len(ev) == 1 {
				if lastEvent != nil && ev[0].ID == lastEvent.ID {
					sleep := es.backoffConfig.Duration()
					es.Logger.Debugw("no new events, backing off", "delaySeconds", sleep)
					time.Sleep(sleep)
					continue
				}
			}

			es.Logger.Debugw("retrieved new events", "count", len(ev))
			ev = removeDuplicates(ev, lastEvent)
			es.Logger.Debugw("remaining new events after filtering out duplicate events", "count", len(ev))

			lastEvent = es.processEvents(ctx, ev, p)
			es.backoffConfig.Reset()
		}
	}
}

// removeDuplicates returns a copy of events with dup element(s) removed
func removeDuplicates(es []AuditEventSummary, dup *AuditEventSummary) []AuditEventSummary {
	cleaned := make([]AuditEventSummary, len(es))
	copy(cleaned, es)

	if dup == nil {
		return cleaned
	}

	for i := range es {
		if es[i].ID == dup.ID {
			// Remove the element at index i from a.
			copy(cleaned[i:], cleaned[i+1:])              // Shift cleaned[i+1:] left one index.
			cleaned[len(cleaned)-1] = AuditEventSummary{} // Erase last element (write zero value).
			cleaned = cleaned[:len(cleaned)-1]            // Truncate slice.
		}
	}
	return cleaned
}

// processEvents sends the given events to the specified processor. Errors from
// the processor will be logged but not returned. There is a risk of poison
// pills here when all events cannot be processed leading to a constant loop in
// the invoking function.
func (es *EventStream) processEvents(ctx context.Context, ev []AuditEventSummary, p processor.Processor) *AuditEventSummary {
	var (
		errCount = 0

		// last successful processed event to track time offset in stream
		lastEvent *AuditEventSummary
	)

	// Horizon events are returned in descending time order
	reverse(ev)

	for i := range ev {
		ce, err := newCloudEvent(ev[i], es.client.Remote())
		if err != nil {
			es.Errorw("skipping event because it could not be converted to CloudEvent format", "event", ev[i], "error", err)
			errCount++
			continue
		}

		es.Infow("invoking processor", "eventID", ce.ID())
		err = p.Process(ctx, *ce)
		if err != nil {
			// retry logic handled inside processor
			es.Errorw("could not process event", "event", ce, "error", err)
			errCount++
			continue
		}
		lastEvent = &ev[i]
	}

	// update metrics
	es.Lock()
	total := *es.stats.EventsTotal + len(ev)
	es.stats.EventsTotal = &total
	errTotal := *es.stats.EventsErr + errCount
	es.stats.EventsErr = &errTotal
	es.Unlock()

	return lastEvent
}

// reverse mutates the given slice and reverses its order
func reverse(ev []AuditEventSummary) {
	for i := len(ev)/2 - 1; i >= 0; i-- {
		opp := len(ev) - 1 - i
		ev[i], ev[opp] = ev[opp], ev[i]
	}
}

func newCloudEvent(event AuditEventSummary, source string) (*cloudevents.Event, error) {
	ce := cloudevents.NewEvent()

	// TODO: revisit CE properties used here
	ce.SetSource(source)
	t := time.Unix(event.Time/1000, 0)
	ce.SetTime(t)
	id := strconv.FormatInt(event.ID, 10)
	ce.SetID(id)

	ce.SetType(convertEventType(event.Type))

	var err error
	err = ce.SetData(events.EventContentType, event)
	if err != nil {
		return nil, errors.Wrap(err, "set CloudEvent data")
	}

	if err = ce.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation for CloudEvent failed")
	}

	return &ce, nil
}

// convertEventType converts a Horizon event type to a normalized cloud event
// type. For example, VLSI_USERLOGGEDIN is converted to
// com.vmware.event.router/horizon.vlsi.userloggedin.v0
func convertEventType(t string) string {
	t = strings.ToLower(t)
	return fmt.Sprintf(eventTypeScheme, events.EventCanonicalType, t)
}

// Shutdown performs a graceful shutdown of the event stream provider
func (es *EventStream) Shutdown(_ context.Context) error {
	if c, ok := es.client.(*horizonClient); ok {
		err := c.logout(context.Background()) // fresh context to avoid canceled err
		if err != nil {
			es.Logger.Warnf("could not log out: %v", err)
		}
	}

	return nil
}
