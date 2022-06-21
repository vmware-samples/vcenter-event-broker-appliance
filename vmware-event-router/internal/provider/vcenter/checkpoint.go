package vcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/events"
)

const (
	defaultCheckpointDir = "checkpoints"
	format               = "cp-%s.json" // cp-<vc_hostname>.json
)

var errInvalidEvent = errors.New("invalid event")

// checkpoint represents a checkpoint object
type checkpoint struct {
	// checkpoint to vc mapping
	VCenter string `json:"vCenter"`
	// last event UUID
	LastEventUUID string `json:"lastEventUUID"`
	// last vCenter event key successfully processed
	LastEventKey int32 `json:"lastEventKey"`
	// last event type, e.g. VmPoweredOffEvent useful for debugging
	LastEventType string `json:"lastEventType"`
	// last vCenter event key timestamp (UTC) successfully processed - used for
	// replaying the event history
	LastEventKeyTimestamp time.Time `json:"lastEventKeyTimestamp"`
	// timestamp (UTC) when this checkpoint was created
	CreatedTimestamp time.Time `json:"createdTimestamp"`
}

// getCheckpoint returns a checkpoint and its full path for the given host and
// directory. If no existing checkpoint is found, an empty checkpoint and
// associated file is created. Thus the checkpoint might be in initialized state
// (i.e. default values) and it is the caller's responsibility to check for
// validity using time.IsZero() on any timestamp.
func getCheckpoint(ctx context.Context, host, dir string) (cp *checkpoint, path string, err error) {
	var skip bool

	file := fileName(host)
	path = fullPath(file, dir)

	f, err := os.Open(path)
	defer func() {
		if f != nil {
			closeWithErrCapture(&err, f, "could not close file")
		}
	}()

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			skip = true
			cp, err = initCheckpoint(ctx, path)
			if err != nil {
				return nil, "", errors.Wrap(err, "could not initialize checkpoint")
			}
		} else {
			return nil, "", errors.Wrap(err, "could not configure checkpointing")
		}
	}

	if !skip {
		cp, err = lastCheckpoint(ctx, f)
		if err != nil {
			return nil, "", errors.Wrap(err, "could not retrieve last checkpoint")
		}
	}

	return cp, path, nil
}

// initCheckpoint creates an empty checkpoint file with default values at the
// given full file path and returns the checkpoint.
func initCheckpoint(_ context.Context, fullPath string) (*checkpoint, error) {
	dir := filepath.Dir(fullPath)

	// create if not exists
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, errors.Wrap(err, "could not create checkpoint directory")
	}

	// create empty checkpoint
	var cp checkpoint
	jsonBytes, err := json.Marshal(cp)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal checkpoint to JSON object")
	}

	err = ioutil.WriteFile(fullPath, jsonBytes, 0o600)
	if err != nil {
		return nil, errors.Wrap(err, "could not write checkpoint file")
	}
	return &cp, nil
}

// lastCheckpoint returns the last checkpoint for the given file
func lastCheckpoint(_ context.Context, file io.Reader) (*checkpoint, error) {
	var cp checkpoint
	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read checkpoint file")
	}

	if err = json.Unmarshal(jsonBytes, &cp); err != nil {
		return nil, errors.Wrap(err, "could not validate last checkpoint")
	}

	return &cp, nil
}

// createCheckpoint creates a checkpoint using the given file, vcenter host name
// and checkpoint timestamp returning the created checkpoint. If lastEvent is
// nil an errInvalidEvent will be returned.
func createCheckpoint(_ context.Context, file io.Writer, vcHost string, last lastEvent, timestamp time.Time) (*checkpoint, error) {
	be := last.baseEvent

	// will panic when the baseEvent value is not pointer
	if be == nil || reflect.ValueOf(be).IsNil() {
		return nil, errInvalidEvent
	}

	createdTime := be.GetEvent().CreatedTime
	eventDetails := events.GetDetails(be)

	cp := checkpoint{
		VCenter:               vcHost,
		LastEventUUID:         last.uuid,
		LastEventKey:          last.key,
		LastEventType:         eventDetails.Name,
		LastEventKeyTimestamp: createdTime,
		CreatedTimestamp:      timestamp,
	}

	b, err := json.Marshal(cp)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal checkpoint to JSON")
	}

	_, err = file.Write(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not write checkpoint")
	}
	return &cp, nil
}

// fileName returns the name of a checkpoint file using the given vCenter host
// as an identifier
func fileName(host string) string {
	return fmt.Sprintf(format, host) // file: eg. cp-<hostname>.json
}

// fullPath returns the full path for the given checkpoint file name and
// directory
func fullPath(file, dir string) string {
	dir = filepath.Clean(dir)
	return filepath.Join(dir, file) // file: e.g. checkpoints/cp-<hostname>.json
}

// closeWithErrCapture runs function and on error return error by argument including the given error (usually
// from caller function).
func closeWithErrCapture(err *error, closer io.Closer, errMsg string) {
	*err = errors.Wrapf(closer.Close(), errMsg)
}
