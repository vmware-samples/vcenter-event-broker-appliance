package tags

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	vsphere "github.com/embano1/vsphere/client"
	"github.com/embano1/vsphere/logger"
	"github.com/google/uuid"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/singleflight"
	"gotest.tools/v3/assert"
)

func newVmPoweredOnEvent() types.BaseEvent {
	return &types.VmPoweredOnEvent{
		VmEvent: types.VmEvent{
			Event: types.Event{
				Key:         1,
				ChainId:     1,
				CreatedTime: time.Now(),
				UserName:    "administrator",
				Vm: &types.VmEventArgument{
					Vm: types.ManagedObjectReference{
						Type:  "VirtualMachine",
						Value: "vm-1",
					},
				},
			},
		},
	}
}

func newCloudEvent(t *testing.T, data interface{}) ce.Event {
	t.Helper()
	e := ce.NewEvent()
	e.SetID(uuid.New().String())
	e.SetType("test.event.v0")
	e.SetSource("test.source")

	err := e.SetData(ce.ApplicationJSON, data)
	assert.NilError(t, err)

	return e
}

func TestSyncer_Run(t *testing.T) {
	t.Run("returns when context cancelled", func(t *testing.T) {
		ceClient, _ := test.NewMockReceiverClient(t, 1)
		s := &Syncer{
			ce:         ceClient,
			serializer: &singleflight.Group{},
		}

		ctx := logger.Set(context.Background(), zaptest.NewLogger(t))
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
		defer cancel()

		err := s.Run(ctx)
		assert.ErrorType(t, err, context.DeadlineExceeded)
	})
}

func TestSyncer_handler(t *testing.T) {
	type wantError struct {
		message string
		code    int
	}

	testCases := []struct {
		name      string
		event     ce.Event
		wantError wantError
	}{
		{
			name:      "cloudevent data is not vsphere event",
			event:     newCloudEvent(t, `{"hello":"world"}`),
			wantError: wantError{message: "could not marshal", code: 400},
		}, {
			name:      "cloudevent data is not tagging event",
			event:     newCloudEvent(t, newVmPoweredOnEvent()),
			wantError: wantError{message: "could not read object", code: 400},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := logger.Set(context.Background(), zaptest.NewLogger(t))

			s := &Syncer{}
			err := s.handler(ctx, tc.event)
			assert.ErrorType(t, err, &http.Result{})
			assert.Equal(t, err.(*http.Result).StatusCode, tc.wantError.code)
		})
	}
}

func TestSyncer_syncTags(t *testing.T) {
	testCases := []struct {
		name    string
		object  string            // vm
		tagMap  map[string]string // category/tag mappings
		nsxMock *nsxAPIMock       // mock NSX API responses
		wantErr string
	}{
		{
			name:    "fails when vm object cannot be found",
			object:  "notexist",
			nsxMock: &nsxAPIMock{},
			wantErr: "not found",
		},
		{
			name:    "fails on nsx http 500 error",
			object:  "DC0_H0_VM0",
			nsxMock: &nsxAPIMock{codes: []int{500}},
			wantErr: "500",
		},
		{
			name:   "successfully synchronizes tags",
			object: "DC0_H0_VM0",
			tagMap: map[string]string{
				"category1": "tag1",
				"category2": "tag2",
			},
			nsxMock: &nsxAPIMock{codes: []int{200}},
			wantErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			simulator.Run(func(ctx context.Context, client *vim25.Client) error {
				ctx = logger.Set(ctx, zaptest.NewLogger(t))

				tc.nsxMock.t = t
				mockHttp := nethttp.DefaultClient
				mockHttp.Transport = tc.nsxMock

				nsxCfg := nsxt.Configuration{
					SkipSessionAuth:      true,
					RetriesConfiguration: nsxt.ClientRetriesConfiguration{}, // disable retries
					HTTPClient:           mockHttp,
				}

				nsxClient, err := nsxt.NewAPIClient(&nsxCfg)
				assert.NilError(t, err)

				govm := govmomi.Client{
					Client:         client,
					SessionManager: session.NewManager(client),
				}

				rc := rest.NewClient(client)
				err = rc.Login(ctx, url.UserPassword("user", "pass"))
				assert.NilError(t, err)
				tm := tags.NewManager(rc)

				s := &Syncer{
					vsphere: &vsphere.Client{
						SOAP: &govm,
						REST: rc,
						Tags: tm,
					},
					nsx: nsxClient,
				}

				var wantID string

				// only required when we don't expect error
				if tc.wantErr == "" {
					ref, err := getVmRef(ctx, s.vsphere.SOAP.Client, tc.object)
					assert.NilError(t, err)
					wantID, err = getInstancedID(ctx, s.vsphere.SOAP.Client, ref)
					assert.NilError(t, err)
					attachTags(t, ctx, s.vsphere.Tags, ref, tc.tagMap)
				}

				_, err = s.syncVmTags(ctx, tc.object, time.Second)()

				if tc.wantErr != "" {
					assert.ErrorContains(t, err, tc.wantErr)
				} else {
					assert.NilError(t, err)
					assert.Equal(t, tc.nsxMock.tagReq.ExternalId, wantID)
					assert.Equal(t, len(tc.tagMap), len(tc.nsxMock.tagReq.Tags))
				}

				return nil
			})
		})
	}
}

type nsxAPIMock struct {
	t     *testing.T
	codes []int // response codes/count

	sync.Mutex
	counter int

	tagReq manager.VirtualMachineTagUpdate // stores last syncer nsx tag request
}

func (nsx *nsxAPIMock) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, req.Body)
	assert.NilError(nsx.t, err)

	var tagReq manager.VirtualMachineTagUpdate
	err = json.Unmarshal(buf.Bytes(), &tagReq)
	assert.NilError(nsx.t, err)

	nsx.Lock()
	defer func() {
		nsx.counter++
		_ = req.Body.Close()
		nsx.Unlock()
	}()

	nsx.tagReq = tagReq

	code := nsx.codes[nsx.counter]
	return &nethttp.Response{
		StatusCode: code,
		Status:     fmt.Sprintf("%d", code),
		Request:    req.Clone(context.TODO()),
	}, nil
}

func attachTags(t *testing.T, ctx context.Context, tm *tags.Manager, ref types.ManagedObjectReference, mappings map[string]string) {
	t.Helper()

	var tagIDs []string
	for cat, tag := range mappings {
		newCat := tags.Category{
			Name:        cat,
			Description: cat,
			// Cardinality: "",
		}
		catID, err := tm.CreateCategory(ctx, &newCat)
		assert.NilError(t, err)

		newTag := tags.Tag{
			Description: tag,
			Name:        tag,
			CategoryID:  catID,
		}
		tagID, err := tm.CreateTag(ctx, &newTag)
		assert.NilError(t, err)

		tagIDs = append(tagIDs, tagID)

	}
	err := tm.AttachMultipleTagsToObject(ctx, tagIDs, ref)
	assert.NilError(t, err)
}
