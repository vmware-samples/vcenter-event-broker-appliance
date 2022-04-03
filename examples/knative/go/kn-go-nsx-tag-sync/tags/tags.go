package tags

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"reflect"
	"sync"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	vsphere "github.com/embano1/vsphere/client"
	"github.com/embano1/vsphere/logger"
	"github.com/kelseyhightower/envconfig"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

// Config configures the syncer
type Config struct {
	// cloudevents listener
	Port  int  `envconfig:"PORT" default:"8080"` // knative injected
	Debug bool `envconfig:"DEBUG" default:"false"`

	// NSX
	NSXAddress    string `envconfig:"NSX_URL" required:"true"`
	NSXInsecure   bool   `envconfig:"NSX_INSECURE" default:"false"`
	NSXSecretPath string `envconfig:"NSX_SECRET_PATH" default:"/var/bindings/nsx"`

	// vCenter
	vsphere.Config
}

// Syncer synchronizes vSphere tags to nsx
type Syncer struct {
	vsphere *vsphere.Client
	ce      ce.Client

	sessionLock sync.RWMutex // periodically recreate nsx session
	nsx         *nsxt.APIClient

	serializer *singleflight.Group // dedupe concurrent operations for same vm object
}

// NewSyncer returns an initialized syncer configured via environment variables
func NewSyncer(ctx context.Context) (*Syncer, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("process environment variables: %w", err)
	}

	log := logger.Get(ctx)
	log.Info("connecting to vcenter", zap.String("host", cfg.Address))
	if cfg.Insecure {
		log.Warn("using potentially insecure connection to vcenter", zap.Bool("insecure", cfg.Insecure))
	}

	var s Syncer
	vc, err := vsphere.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create vsphere client: %w", err)
	}
	s.vsphere = vc

	nsx, err := newNSXClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create nsx client: %w", err)
	}

	s.nsx = nsx

	s.serializer = &singleflight.Group{}
	ceClient, err := ce.NewClientHTTP(http.WithPort(cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("create cloudevents http client: %w", err)
	}
	s.ce = ceClient

	return &s, nil
}

// Run runs the syncer
func (s *Syncer) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.ce.StartReceiver(egCtx, s.handler)
	})

	// nsx session handling
	eg.Go(func() error {
		ticker := time.NewTicker(nsxSessionRefresh)
		defer ticker.Stop()

		for {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			case <-ticker.C:
				reauth := func() error {
					logger.Get(egCtx).Debug("attempting to reauthenticate nsx session")
					s.sessionLock.Lock()
					defer s.sessionLock.Unlock()

					// note: could block up to nsxAPITimeout
					if err := nsxt.GetDefaultHeaders(s.nsx); err != nil {
						return fmt.Errorf("reauthenticate nsx session: %w", err)
					}
					logger.Get(egCtx).Debug("successfully reauthenticated nsx session")

					return nil
				}()

				if err := reauth; err != nil {
					return err
				}
			}
		}
	})

	eg.Go(func() error {
		<-egCtx.Done()
		logger.Get(egCtx).Info("received shutdown signal", zap.String("signal", egCtx.Err().Error()))
		return nil
	})

	return eg.Wait()
}

func (s *Syncer) handler(ctx context.Context, event ce.Event) error {
	log := logger.Get(ctx).With(zap.String("eventID", event.ID()))
	log.Debug("received event", zap.Any("event", event))

	var vevent types.EventEx
	if err := event.DataAs(&vevent); err != nil {
		log.Error("could not marshal event to eventex", zap.Error(err))
		return http.NewResult(nethttp.StatusBadRequest,
			"could not marshal cloudevent event data to vsphere event (eventID: %d)",
			event.ID(),
		)
	}

	var object string
	for _, arg := range vevent.Arguments {
		if arg.Key == "Object" {
			if o, ok := arg.Value.(string); ok {
				object = o
				break
			} else {
				valueType := reflect.TypeOf(arg.Value).String()
				log.Error("could not convert eventex argument value to string",
					zap.Any("argumentObject", arg),
					zap.String("valueType", valueType),
				)
				return http.NewResult(nethttp.StatusBadRequest,
					"could not read object value from event arguments (eventID: %d)",
					event.ID(),
				)
			}
		}
	}

	if object == "" {
		log.Error("event did not contain object key", zap.Any("event", event))
		return http.NewResult(nethttp.StatusBadRequest,
			"could not read object value from event arguments (eventID: %d)",
			event.ID(),
		)
	}

	ctx = logger.Set(ctx, log)
	nsxTimeout := time.Second * 3
	_, err, shared := s.serializer.Do(object, s.syncVmTags(ctx, object, nsxTimeout))
	if shared {
		log.Info("serialized and deduplicated concurrent calls to object ", zap.String("object", object))
	}

	if err != nil {
		// either tag not on a vm object or vm already removed from inventory
		var nfe *find.NotFoundError
		if errors.As(err, &nfe) {
			log.Warn("ignoring object", zap.String("object", object), zap.Error(err))

			// return 400 instead of 404 because some brokers retry on 404
			return http.NewResult(nethttp.StatusBadRequest,
				"could not synchronize tags for object %q (eventID: %d)",
				object,
				event.ID(),
			)
		}

		log.Error("could not synchronize vm tags", zap.String("object", object), zap.Error(err))
		return http.NewResult(nethttp.StatusInternalServerError,
			"could not synchronize tags for vm %q (eventID: %d)",
			object,
			event.ID(),
		)
	}

	return nil
}

func (s *Syncer) syncVmTags(ctx context.Context, vm string, nsxTimeout time.Duration) func() (interface{}, error) {
	return func() (interface{}, error) {
		log := logger.Get(ctx)
		log.Debug("retrieving vm managed object reference", zap.String("object", vm))

		ref, err := getVmRef(ctx, s.vsphere.SOAP.Client, vm)
		if err != nil {
			return nil, fmt.Errorf("find vm reference for object %q: %w", vm, err)
		}

		log.Debug("retrieving vm instance id", zap.Any("ref", ref))
		id, err := getInstancedID(ctx, s.vsphere.SOAP.Client, ref)
		if err != nil {
			return nil, fmt.Errorf("retrieve vm instance id for ref %q: %w", ref, err)
		}

		log.Debug("retrieving vm tags", zap.Any("ref", ref))
		attachedTags, err := s.vsphere.Tags.ListAttachedTags(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("list attached tags: %w", err)
		}

		req := manager.VirtualMachineTagUpdate{
			ExternalId: id,
			Tags:       make([]common.Tag, len(attachedTags)),
		}

		for idx, t := range attachedTags {
			details, err := s.vsphere.Tags.GetTag(ctx, t)
			if err != nil {
				return nil, fmt.Errorf("get tag %q: %w", t, err)
			}

			categoryDetails, err := s.vsphere.Tags.GetCategory(ctx, details.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("get category %q: %w", details.CategoryID, err)
			}

			req.Tags[idx] = common.Tag{
				Scope: categoryDetails.Name,
				Tag:   details.Name,
			}
		}

		// use explicitly provided timeout here to return fast from this singleflight
		// routine and reduce likelihood of stale tag information during concurrent
		// (blocked) operations on the same object (key)
		//
		// note: in case timeout fires before nsx ack, singleflight caller returns HTTP
		// 500 due to ctx.Error() in the event handler, causing event sender (broker)
		// typically to retry
		ctx, cancel := context.WithTimeout(ctx, nsxTimeout)
		defer cancel()
		log.Info("updating virtual machine tags in nsx", zap.Any("request", req))

		updateTags := func() error {
			s.sessionLock.RLock()
			defer s.sessionLock.RUnlock()
			if _, err = s.nsx.FabricApi.UpdateVirtualMachineTagsUpdateTags(ctx, req); err != nil {
				return err
			}
			return nil
		}

		if err = updateTags(); err != nil {
			return nil, fmt.Errorf("update virtual machine tags in nsx: %w", err)
		}

		return nil, nil
	}
}
