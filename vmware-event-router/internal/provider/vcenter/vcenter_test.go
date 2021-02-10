// +build unit

package vcenter

import (
	"context"
	"fmt"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"knative.dev/pkg/logging"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

func TestEventStream_stream(t *testing.T) {
	type args struct {
		enableCheckpoint bool
	}
	type fields struct {
		begin time.Time
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    int
		wantErr error
	}{
		{
			name: "expect no events",
			fields: fields{
				begin: time.Time{}, // will be "now" in test
			},
			args: args{
				enableCheckpoint: false,
			},
			want:    0,
			wantErr: context.Canceled,
		},
		{
			name: "expect all events",
			fields: fields{
				begin: time.Now().UTC().Add(time.Hour * -1),
			},
			args: args{
				enableCheckpoint: false,
			},
			want:    26, // current number returned by default VPX simulator model
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simulator.Run(func(simCtx context.Context, client *vim25.Client) error {
				logger := zaptest.NewLogger(t).Sugar()
				simCtx = logging.WithLogger(simCtx, logger)

				ctx, cancel := context.WithCancel(simCtx)
				defer cancel()

				c := govmomi.Client{
					Client:         client,
					SessionManager: session.NewManager(client),
				}

				vc := &EventStream{
					client:     c,
					Logger:     logger,
					checkpoint: tt.args.enableCheckpoint,
					stats: metrics.EventStats{
						EventsTotal: new(int),
						EventsErr:   new(int),
						EventsSec:   new(float64),
					},
				}

				begin := time.Now().UTC()
				if !tt.fields.begin.IsZero() {
					begin = tt.fields.begin
				}

				coll, err := newHistoryCollector(ctx, client, &begin)
				if err != nil {
					t.Fatalf("create history collector: %v", err)
				}

				proc := fakeProcessor{
					expect: tt.want,
					doneCh: make(chan struct{}),
				}

				eg, egCtx := errgroup.WithContext(ctx)

				// stream
				eg.Go(func() error {
					return vc.stream(egCtx, &proc, coll, tt.args.enableCheckpoint)
				})

				// give streamer a bit time to establish vc connection
				time.Sleep(time.Second)

				// processor
				eg.Go(func() error {
					if proc.expect == 0 {
						cancel()
						return nil
					}

					select {
					case <-egCtx.Done():
						return egCtx.Err()
					case <-proc.doneCh:
						cancel()
						return nil
					}
				})

				// cancel early in case of issues
				eg.Go(func() error {
					select {
					case <-time.After(time.Second * 3):
						return fmt.Errorf("timed out")
					case <-egCtx.Done():
						return egCtx.Err()
					}
				})

				if err = eg.Wait(); err != tt.wantErr {
					t.Errorf("stream() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.want != proc.got {
					t.Errorf("stream() events got = %v, want %v", proc.got, tt.want)
				}

				return nil
			})
		})
	}
}

type fakeProcessor struct {
	got    int
	expect int
	doneCh chan struct{}
}

func (f *fakeProcessor) Process(ctx context.Context, ce cloudevents.Event) error {
	logging.FromContext(ctx).Debugf("processing event: %s", ce.String())
	f.got++
	if f.expect == f.got {
		close(f.doneCh)
	}
	return nil
}

func (f *fakeProcessor) PushMetrics(_ context.Context, _ metrics.Receiver) {
}

func (f *fakeProcessor) Shutdown(_ context.Context) error {
	return nil
}
