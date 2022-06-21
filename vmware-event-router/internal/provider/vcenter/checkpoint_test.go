//go:build unit

package vcenter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

const (
	vcHost = "vcenter-01.corp.local"
)

var (
	now              = time.Now().UTC()
	lastEventUUID    = "e5b89c60-d292-44cb-b5f0-7fef6d0e5fb4"
	lastEventKey     = int32(1234)
	lastEventType    = "VmPoweredOffEvent"
	lastEventKeyTime = now.Add(time.Minute * -30) // 30min in past
	validCheckpoint  = checkpoint{
		VCenter:               vcHost,
		LastEventUUID:         lastEventUUID,
		LastEventKey:          lastEventKey,
		LastEventType:         lastEventType,
		LastEventKeyTimestamp: lastEventKeyTime,
		CreatedTimestamp:      now,
	}
)

// tempDir creates a temporary directory and returns its path and a function to
// delete the directory
func tempDir(t *testing.T) (path string, cleanup func()) {
	t.Helper()

	const tmpDirPrefix = "event-router-tests"
	tmpDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		t.Fatalf("could not create temporary directory: %v", err)
	}

	return tmpDir, func() {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("could not remove temporary directory: %v", err)
		}
	}
}

// createTempCheckpoint returns a temporary and valid checkpoint file, its name and
// directory
func createTempCheckpoint(t *testing.T) (*os.File, string, string) {
	t.Helper()

	cpDir, removeDirFn := tempDir(t)
	t.Cleanup(removeDirFn)

	file := fileName(vcHost)
	path := fullPath(file, cpDir)

	f, err := os.Create(path)
	if err != nil {
		t.Errorf("could not create checkpoint file: %v", err)
	}

	jsonByte, err := json.Marshal(validCheckpoint)
	if err != nil {
		t.Errorf("could not marshal to JSON: %v", err)
	}

	_, err = f.Write(jsonByte)
	if err != nil {
		t.Errorf("could not write checkpoint file: %v", err)
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		t.Errorf("could not reset checkpoint file: %v", err)
	}
	return f, file, cpDir
}

func Test_checkpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jsonValid, err := json.Marshal(validCheckpoint)
	if err != nil {
		t.Errorf("could not marshal checkpoint to JSON: %v", err)
	}

	type args struct {
		ctx       context.Context
		vcHost    string
		lastEvent lastEvent
	}
	tests := []struct {
		name     string
		args     args
		wantFile string
		wantCP   *checkpoint
		key      int32
		wantErr  bool
	}{
		{
			name: "valid lastEvent",
			args: args{
				ctx:    ctx,
				vcHost: vcHost,
				lastEvent: lastEvent{
					baseEvent: &types.VmPoweredOffEvent{
						VmEvent: types.VmEvent{
							Event: types.Event{
								Key:         lastEventKey,
								CreatedTime: lastEventKeyTime,
							},
						},
					},
					uuid: lastEventUUID,
					key:  lastEventKey,
				},
			},
			wantFile: string(jsonValid),
			wantCP:   &validCheckpoint,
			key:      lastEventKey,
			wantErr:  false,
		},
		{
			name: "invalid lastEvent (event is nil)",
			args: args{
				ctx:       ctx,
				vcHost:    vcHost,
				lastEvent: lastEvent{baseEvent: nil},
			},
			wantFile: "",
			wantCP:   nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &bytes.Buffer{}
			cp, err := createCheckpoint(tt.args.ctx, file, tt.args.vcHost, tt.args.lastEvent, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile := file.String(); gotFile != tt.wantFile {
				t.Errorf("checkpoint() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if !reflect.DeepEqual(cp, tt.wantCP) {
				t.Errorf("checkpoint() gotCheckpoint = %v, want %v", cp, tt.wantCP)
			}
			if cp != nil {
				if cp.LastEventKey != tt.key {
					t.Errorf("checkpoint() gotKey = %v, want %v", cp.LastEventKey, tt.key)
				}
			}
		})
	}
}

func Test_fullPath(t *testing.T) {
	const (
		fileName           = "cp-vcenter-01-prod.json"
		customFullPathUnix = "/tmp/router/cps"
		customRelPathUnix  = "./mycheckpoints/"
	)

	type args struct {
		file string
		dir  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default relative path",
			args: args{
				file: fileName,
				dir:  defaultCheckpointDir,
			},
			want: "checkpoints/" + fileName,
		},
		{
			name: "custom full path UNIX",
			args: args{
				file: fileName,
				dir:  customFullPathUnix,
			},
			want: customFullPathUnix + "/" + fileName,
		},
		{
			name: "custom relative path",
			args: args{
				file: fileName,
				dir:  customRelPathUnix,
			},
			want: strings.Trim(customRelPathUnix, "./") + "/" + fileName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fullPath(tt.args.file, tt.args.dir); got != tt.want {
				t.Errorf("fullPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lastCheckpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, _, _ := createTempCheckpoint(t)
	defer func() {
		err := valid.Close()
		if err != nil {
			t.Errorf("could not close file: %v", err)
		}
	}()

	empty := bytes.NewBufferString(`{
  "vCenter": "",
  "lastEventKey": 0,
  "lastEventKeyTimestamp": "0001-01-01T00:00:00Z",
  "checkpointTimestamp": "0001-01-01T00:00:00Z"
}
`)

	invalid := bytes.NewBufferString("{")

	type args struct {
		ctx  context.Context
		file io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *checkpoint
		wantErr bool
	}{
		{
			name: "valid checkpoint",
			args: args{
				ctx:  ctx,
				file: valid,
			},
			want:    &validCheckpoint,
			wantErr: false,
		},
		{
			name: "empty checkpoint",
			args: args{
				ctx:  ctx,
				file: empty,
			},
			want: &checkpoint{
				VCenter:               "",
				LastEventKey:          0,
				LastEventKeyTimestamp: time.Time{},
				CreatedTimestamp:      time.Time{},
			},
			wantErr: false,
		},
		{
			name: "invalid checkpoint",
			args: args{
				ctx:  ctx,
				file: invalid,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lastCheckpoint(tt.args.ctx, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("lastCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lastCheckpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCheckpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, name, dir := createTempCheckpoint(t)
	defer func() {
		err := valid.Close()
		if err != nil {
			t.Errorf("could not close file: %v", err)
		}
	}()

	emptyDir, cleanup := tempDir(t)
	defer cleanup()

	var (
		emptyCpDirWithSlash       = emptyDir + "/"
		emptyCpDirWithDoubleSlash = emptyDir + "//"
		emptyCp                   checkpoint
	)

	type args struct {
		ctx  context.Context
		host string
		dir  string
	}

	tests := []struct {
		name     string
		args     args
		want     *checkpoint
		fullPath string
		wantErr  bool
	}{
		{
			name: "existing checkpoint",
			args: args{
				ctx:  ctx,
				host: vcHost,
				dir:  dir,
			},
			want:     &validCheckpoint,
			fullPath: fullPath(name, dir),
			wantErr:  false,
		},
		{
			name: "not existing checkpoint",
			args: args{
				ctx:  ctx,
				host: "host-02",
				dir:  emptyDir,
			},
			want:     &emptyCp,
			fullPath: fullPath("cp-host-02.json", emptyDir),
			wantErr:  false,
		},
		{
			name: "not existing checkpoint with trailing slash",
			args: args{
				ctx:  ctx,
				host: "host-02",
				dir:  emptyCpDirWithSlash,
			},
			want:     &emptyCp,
			fullPath: fullPath("cp-host-02.json", emptyDir),
			wantErr:  false,
		},
		{
			name: "not existing checkpoint with double trailing slash",
			args: args{
				ctx:  ctx,
				host: "host-03",
				dir:  emptyCpDirWithDoubleSlash,
			},
			want:     &emptyCp,
			fullPath: fullPath("cp-host-03.json", emptyDir),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp, file, err := getCheckpoint(tt.args.ctx, tt.args.host, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCheckpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(cp, tt.want) {
				t.Errorf("getCheckpoint() got = %v, want %v", cp, tt.want)
			}
			if file != tt.fullPath {
				t.Errorf("getCheckpoint() file = %v, want %v", file, tt.fullPath)
			}
		})
	}
}
