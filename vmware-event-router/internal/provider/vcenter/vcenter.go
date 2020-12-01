package vcenter

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/embano1/waitgroup"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"

	"github.com/jpillora/backoff"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider"
)

const (
	defaultPollFrequency  = time.Second
	eventsPageMax         = 100 // events per page from history collector
	checkpointInterval    = 5 * time.Second
	checkpointMaxEventAge = time.Hour       // limit event replay time window to max
	waitShutdown          = 5 * time.Second // wait for processing to finish during shutdown
)

// EventStream handles the connection to the vCenter events API
type EventStream struct {
	client govmomi.Client
	logger.Logger
	checkpoint    bool
	checkpointDir string

	wg waitgroup.WaitGroup // shutdown handling

	sync.RWMutex
	stats metrics.EventStats
}

type lastEvent struct {
	baseEvent types.BaseEvent
	uuid      string
	key       int32
}

// assert we implement Provider interface
var _ provider.Provider = (*EventStream)(nil)

// NewEventStream returns a vCenter event stream manager for a given
// configuration and metrics server
func NewEventStream(ctx context.Context, cfg *config.ProviderConfigVCenter, ms metrics.Receiver, log logger.Logger, opts ...Option) (*EventStream, error) {
	if cfg == nil {
		return nil, errors.New("vCenter configuration must be provided")
	}

	var vc EventStream

	parsedURL, err := soap.ParseURL(cfg.Address)
	if err != nil {
		return nil, errors.Wrap(err, "parsing vCenter URL")
	}

	// TODO: only supporting basic auth against vCenter for now
	if cfg.Auth == nil || cfg.Auth.BasicAuth == nil {
		return nil, fmt.Errorf("invalid %s credentials: username and password must be set", config.BasicAuth)
	}

	username := cfg.Auth.BasicAuth.Username
	password := cfg.Auth.BasicAuth.Password
	parsedURL.User = url.UserPassword(username, password)

	client, err := govmomi.NewClient(ctx, parsedURL, cfg.InsecureSSL)
	if err != nil {
		return nil, errors.Wrap(err, "create vCenter client")
	}

	vcLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		prov := strings.ToUpper(string(config.ProviderVCenter))
		vcLog = zapSugared.Named(fmt.Sprintf("[%s]", prov))
	}

	vc.Logger = vcLog
	vc.client = *client
	vc.checkpoint = cfg.Checkpoint
	vc.checkpointDir = cfg.CheckpointDir

	// apply options (overwrite any defaults)
	for _, opt := range opts {
		opt(&vc)
	}

	// seed the metrics stats
	vc.stats = metrics.EventStats{
		Provider:    string(config.ProviderVCenter),
		Type:        config.EventProvider,
		Address:     cfg.Address,
		Started:     time.Now().UTC(),
		EventsTotal: new(int),
		EventsErr:   new(int),
		EventsSec:   new(float64),
	}

	go vc.PushMetrics(ctx, ms)
	return &vc, nil
}

// Stream is the main logic, blocking to receive and handle events from vCenter
func (vc *EventStream) Stream(ctx context.Context, p processor.Processor) error {
	var (
		begin *time.Time
		cp    *checkpoint
		path  string
		err   error
	)

	// begin of event stream defaults to current vCenter time (UTC)
	begin, err = methods.GetCurrentTime(ctx, vc.client)
	if err != nil {
		return errors.Wrap(err, "get current time from vCenter")
	}

	// configure checkpointing and retrieve last checkpoint, if any
	switch vc.checkpoint {
	case true:
		vc.Info("enabling checkpoints and checking for existing checkpoint")
		host := vc.client.URL().Hostname()

		dir := defaultCheckpointDir
		if vc.checkpointDir != "" {
			dir = vc.checkpointDir
		}

		cp, path, err = getCheckpoint(ctx, host, dir)
		if err != nil {
			return errors.Wrap(err, "get checkpoint")
		}

		// if the timestamp is valid set begin to last checkpoint
		ts := cp.LastEventKeyTimestamp
		if !ts.IsZero() {
			vc.Infow("found existing and valid checkpoint", "path", path)
			// perform boundary check
			maxTS := begin.Add(checkpointMaxEventAge * -1)
			if maxTS.Unix() > ts.Unix() {
				begin = &maxTS
				vc.Warnw("last event timestamp in checkpoint is older than configured maximum", "maxTimestamp", checkpointMaxEventAge.String())
				vc.Warnw("setting begin of event stream", "beginTimestamp", begin.String())
			} else {
				begin = &ts
				vc.Infow("setting begin of event stream", "beginTimestamp", begin.String(), "eventKey", cp.LastEventKey)
			}
		} else {
			vc.Info("no valid checkpoint found")
			vc.Infow("empty checkpoint created", "path", path)
			vc.Infow("setting begin of event stream", "beginTimestamp", begin.String())
		}

	case false:
		vc.Infow("checkpointing disabled, setting begin of event stream", "beginTimestamp", begin.String())
	}

	ec, err := newHistoryCollector(ctx, vc.client.Client, begin)
	if err != nil {
		return errors.Wrap(err, "create event history collector")
	}

	defer func() {
		// use new ctx bc current might be cancelled
		if ctx.Err() != nil {
			ctx = context.Background()
		}
		_ = ec.Destroy(ctx) // ignore any err
	}()

	vc.wg.Add(1)
	defer vc.wg.Done()
	return vc.stream(ctx, p, ec, vc.checkpoint)
}

func (vc *EventStream) stream(ctx context.Context, p processor.Processor, collector *event.HistoryCollector, enableCheckpoint bool) error {
	// event poll ticker
	pollTick := time.NewTicker(defaultPollFrequency)
	defer pollTick.Stop()

	// create checkpoint ticker only if needed
	var cpTick <-chan time.Time = nil
	if enableCheckpoint {
		cpTicker := time.NewTicker(checkpointInterval)
		cpTick = cpTicker.C
		defer cpTicker.Stop()
	}

	var (
		last      *lastEvent // last processed event
		lastCpKey int32      // last event key in checkpoint
		bOff      = backoff.Backoff{
			Factor: 2,
			Jitter: false,
			Min:    time.Second,
			Max:    5 * time.Second,
		}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		// 	there is a small chance (timing and channel handling) that we received
		// 	event(s) and crashed before creating the first checkpoint. at-least-once
		// 	would be violated because we come back with an empty initialized checkpoint.
		// 	TODO: we could force a checkpoint after the first event to reduce the likelihood
		case <-cpTick:

			// skip if checkpoint channel fires before first event or no new events received
			// since last checkpoint
			if last == nil || (last.key == lastCpKey) {
				vc.Debug("no new events, skipping checkpoint")
				continue
			}

			host := vc.client.URL().Hostname()
			f := fileName(host)

			dir := defaultCheckpointDir
			if vc.checkpointDir != "" {
				dir = vc.checkpointDir
			}
			path := fullPath(f, dir)

			// always create/overwrite (existing) checkpoint
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			if err != nil {
				return errors.Wrap(err, "create checkpoint file")
			}

			cp, err := createCheckpoint(ctx, file, host, *last, time.Now().UTC())
			if err != nil {
				return errors.Wrap(err, "create checkpoint")
			}
			lastCpKey = cp.LastEventKey

			err = file.Close()
			if err != nil {
				return errors.Wrap(err, "close checkpoint file")
			}

			vc.Infow("created checkpoint", "path", path, "eventKey", lastCpKey)

		case <-pollTick.C:
			baseEvents, err := collector.ReadNextEvents(ctx, eventsPageMax)
			// TODO: handle error without returning?
			if err != nil {
				return errors.Wrap(err, "retrieve events")
			}

			if len(baseEvents) == 0 {
				sleep := bOff.Duration()
				vc.Debugw("no new events, backing off", "delaySeconds", sleep)
				time.Sleep(sleep)
				continue
			}

			last = vc.processEvents(ctx, baseEvents, p)
			bOff.Reset()
		}
	}
}

// processEvents processes events from vcenter serially, i.e. in order, invoking
// the supplied processor. Errors are logged and tracked in the metric stats.
// The last event processed, including those returning with error, is returned.
func (vc *EventStream) processEvents(ctx context.Context, baseEvents []types.BaseEvent, p processor.Processor) *lastEvent {
	var (
		errCount int
		last     *lastEvent
	)

	host := vc.client.URL().String()

	for _, e := range baseEvents {
		ce, err := events.NewCloudEvent(e, host)
		if err != nil {
			vc.Errorw("skipping event because it could not be converted to CloudEvent format", "event", e, "error", err)
			errCount++
			continue
		}

		vc.Infow("invoking processor", "eventID", ce.ID())
		err = p.Process(ctx, *ce)
		if err != nil {
			// retry logic handled inside processor
			vc.Errorw("could not process event", "event", ce, "error", err)
			errCount++
		}
		last = &lastEvent{
			baseEvent: e,
			uuid:      ce.ID(),
			key:       e.GetEvent().Key,
		}
	}

	// update metrics
	vc.Lock()
	total := *vc.stats.EventsTotal + len(baseEvents)
	vc.stats.EventsTotal = &total
	errTotal := *vc.stats.EventsErr + errCount
	vc.stats.EventsErr = &errTotal
	vc.Unlock()

	return last
}

// Shutdown closes the underlying connection to vCenter
func (vc *EventStream) Shutdown(ctx context.Context) error {
	vc.Logger.Infof("attempting graceful shutdown")
	if err := vc.wg.WaitTimeout(waitShutdown); err != nil {
		return errors.Wrap(err, "shutdown")
	}

	// create new ctx in case current already cancelled
	if ctx.Err() != nil {
		ctx = context.Background()
	}
	return errors.Wrap(vc.client.Logout(ctx), "logout from vCenter") // err == nil if logout was successful
}

// PushMetrics pushes metrics to the configured metrics receiver
func (vc *EventStream) PushMetrics(ctx context.Context, ms metrics.Receiver) {
	ticker := time.NewTicker(metrics.PushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vc.Lock()
			eventsSec := math.Round((float64(*vc.stats.EventsTotal)/time.Since(vc.stats.Started).Seconds())*100) / 100 // 0.2f syntax
			vc.stats.EventsSec = &eventsSec
			ms.Receive(&vc.stats)
			vc.Unlock()
		}
	}
}

func newHistoryCollector(ctx context.Context, vcClient *vim25.Client, begin *time.Time) (*event.HistoryCollector, error) {
	mgr := event.NewManager(vcClient)
	root := vcClient.ServiceContent.RootFolder

	// configure the event stream filter (begin of stream)
	filter := types.EventFilterSpec{
		// EventTypeId: []string{...}, // only stream specific types, e.g. VmEvent
		Entity: &types.EventFilterSpecByEntity{
			Entity:    root,
			Recursion: types.EventFilterSpecRecursionOptionAll,
		},
		Time: &types.EventFilterSpecByTime{
			BeginTime: types.NewTime(*begin),
		},
	}

	return mgr.CreateCollectorForEvents(ctx, filter)
}
