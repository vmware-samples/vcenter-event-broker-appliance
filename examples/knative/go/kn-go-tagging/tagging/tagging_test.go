package tagging

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/vmware/govmomi/simulator"
	_ "github.com/vmware/govmomi/vapi/simulator"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap/zaptest"
	"gotest.tools/assert"
	"knative.dev/pkg/logging"
)

func TestParseEventMoRef(t *testing.T) {
	type vm struct {
		vmType string
		Value  string
	}

	var tests = []struct {
		name      string
		path      string
		expectErr bool
		want      *vm
	}{
		{
			"Test that event is readable",
			"testdata/event.json",
			false,
			&vm{vmType: "VirtualMachine", Value: "vm-10000"},
		},
		{
			"Event should return error if vm type and value are null",
			"testdata/eventErr1.json",
			true,
			nil,
		},
		{
			"Event should return error if vm info is null",
			"testdata/eventErr2.json",
			true,
			nil,
		},
		{
			"Event should return error if vm parent is null",
			"testdata/eventErr3.json",
			true,
			nil,
		},
		{
			"Event should return error if data is null",
			"testdata/eventErr4.json",
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := jsonFileToEvent(tt.path)
			moRef, err := parseEventMoRef(e)

			if tt.expectErr && err == nil {
				t.Error("expected error, but got none")
			}

			if !tt.expectErr && err != nil {
				assert.Equal(t, moRef.Type, tt.want.vmType)
				assert.Equal(t, moRef.Value, tt.want.Value)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	const (
		username = "administrator@vsphere.local"
		password = "pass"
	)

	secretsDir := createSecret(t, username, password)

	t.Run("fail with authentication error", func(t *testing.T) {
		model := simulator.VPX()
		defer model.Remove()

		err := model.Create()
		assert.NilError(t, err, "create vcsim model")

		model.Service.Listen = &url.URL{
			User: url.UserPassword("not-my-username", password),
		}

		simulator.Run(func(ctx context.Context, client *vim25.Client) error {
			defaultEnv := EnvConfig{
				Port:       50001,
				SecretPath: secretsDir,
				Insecure:   true, // vcsim
				VCAddress:  client.URL().String(),
			}

			err = setEnv(defaultEnv)
			assert.NilError(t, err, "set environment variables")

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			ctx = logging.WithLogger(ctx, zaptest.NewLogger(t).Sugar())

			_, err = NewClient(ctx)
			if err != nil {
				msg := "SOAP session manager login: ServerFaultCode: Login failure"
				assert.ErrorContains(t, err, msg)
			}

			return nil
		}, model)
	})

	t.Run("attach a tag to eventing vm", func(t *testing.T) {
		model := simulator.VPX()
		defer model.Remove()
		err := model.Create()
		assert.NilError(t, err, "create vcsim model")
		model.Service.Listen = &url.URL{
			User: url.UserPassword(username, password),
		}

		simulator.Run(func(ctx context.Context, client *vim25.Client) error {
			defaultEnv := EnvConfig{
				Port:       50001,
				SecretPath: secretsDir,
				Insecure:   true, // vcsim
				TagName:    "example-tag",
				TagAction:  "attach",
				VCAddress:  client.URL().String(),
			}
			err := setEnv(defaultEnv)
			assert.NilError(t, err, "set environment variables")

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			ctx = logging.WithLogger(ctx, zaptest.NewLogger(t).Sugar())

			goTagClient, err := NewClient(ctx)
			assert.NilError(t, err, "authenticating")

			testTagID, err := createTag(t, ctx, goTagClient.tagMgr, defaultEnv.TagName)
			assert.NilError(t, err, "creating tag")

			event := jsonFileToEvent("testdata/event.json")
			result := goTagClient.handler(ctx, event)
			assert.NilError(t, result, "expect nil result")

			// Managed Object Reference of the eventing VM
			vmMoRef := types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "vm-56",
			}

			attached, err := isTagAttached(t, ctx, goTagClient.tagMgr, vmMoRef, testTagID)
			assert.NilError(t, err, "checking if tag", defaultEnv.TagName, "is attached to", vmMoRef.Value)
			if !attached {
				t.Errorf("expect tag %q to be attached to %q but it is not", testTagID, vmMoRef.Value)
			}

			return nil
		}, model)
	})

	t.Run("detach a tag from eventing vm", func(t *testing.T) {
		model := simulator.VPX()
		err := model.Create()
		assert.NilError(t, err, "create vcsim model")
		model.Service.Listen = &url.URL{
			User: url.UserPassword(username, password),
		}

		simulator.Run(func(ctx context.Context, client *vim25.Client) error {
			defaultEnv := EnvConfig{
				Port:       50001,
				SecretPath: secretsDir,
				Insecure:   true, // vcsim
				TagName:    "example-tag-to-detach",
				TagAction:  "detach",
				VCAddress:  client.URL().String(),
			}
			err := setEnv(defaultEnv)
			assert.NilError(t, err, "set environment variables")

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			ctx = logging.WithLogger(ctx, zaptest.NewLogger(t).Sugar())

			goTagClient, err := NewClient(ctx)
			assert.NilError(t, err, "authenticating")

			testTagID, err := createTag(t, ctx, goTagClient.tagMgr, defaultEnv.TagName)
			assert.NilError(t, err, "creating tag")

			// Managed Object Reference of the eventing VM
			vmMoRef := types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "vm-60",
			}

			err = goTagClient.tagMgr.AttachTag(ctx, testTagID, vmMoRef)
			assert.NilError(t, err, "attach tag", defaultEnv.TagName, "to", vmMoRef.Value)

			attached, err := isTagAttached(t, ctx, goTagClient.tagMgr, vmMoRef, testTagID)
			assert.NilError(t, err, "checking if tag ", defaultEnv.TagName, "is attached to", vmMoRef.Value)
			if !attached {
				t.Errorf("expected tag %q to be attached to %q, but it is not", testTagID, vmMoRef.Value)
			}

			event := jsonFileToEvent("testdata/event2.json")
			result := goTagClient.handler(ctx, event)
			assert.NilError(t, result, "expect nil result")

			attached, err = isTagAttached(t, ctx, goTagClient.tagMgr, vmMoRef, testTagID)
			assert.NilError(t, err, "checking if tag ", defaultEnv.TagName, "is attached to", vmMoRef.Value)
			if attached {
				t.Errorf("expected tag %q to be detached from %q, but it is not", testTagID, vmMoRef.Value)
			}

			return nil
		}, model)
	})
}

func jsonFileToEvent(path string) event.Event {
	e := event.New(event.CloudEventsVersionV03)

	json, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("err", err)
	}

	err = e.SetData(event.ApplicationJSON, json)
	if err != nil {
		fmt.Println("err", err)
	}

	return e
}

// createSecret returns a directory with username and password files holding
// user/pass credentials
func createSecret(t *testing.T, username, password string) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "k8s-secret")
	assert.NilError(t, err, "create secrets directory")

	t.Cleanup(func() {
		err = os.RemoveAll(dir)
		assert.NilError(t, err, "cleanup temp directory")
	})

	userFile := filepath.Join(dir, "username")
	err = ioutil.WriteFile(userFile, []byte(username), 0444)
	assert.NilError(t, err, "create username secret")

	passFile := filepath.Join(dir, "password")
	err = ioutil.WriteFile(passFile, []byte(password), 0444)
	assert.NilError(t, err, "create password secret")

	return dir
}

// setEnv sets all environment variables defined in env
func setEnv(env EnvConfig) error {
	if err := os.Setenv("PORT", strconv.Itoa(env.Port)); err != nil {
		return err
	}

	if err := os.Setenv("VCENTER_URL", env.VCAddress); err != nil {
		return err
	}

	if err := os.Setenv("VCENTER_INSECURE", strconv.FormatBool(env.Insecure)); err != nil {
		return err
	}

	if err := os.Setenv("VCENTER_SECRET_PATH", env.SecretPath); err != nil {
		return err
	}

	if err := os.Setenv("TAG_NAME", env.TagName); err != nil {
		return err
	}

	if err := os.Setenv("TAG_ACTION", env.TagAction); err != nil {
		return err
	}

	if err := os.Setenv("DEBUG", strconv.FormatBool(env.DebugLogs)); err != nil {
		return err
	}

	return nil
}

// createTags creates the given tag to category mappings and returns a map of
// names to IDs (URNs) for all created tags
func createTag(t *testing.T, ctx context.Context, mgr *tags.Manager, tagName string) (string, error) {
	t.Helper()

	cat := tags.Category{
		Name:        "test-category-1",
		Description: "category used for testing against simulator",
		Cardinality: "MULTIPLE",
	}
	catID, err := mgr.CreateCategory(ctx, &cat)
	if err != nil {
		return "", err
	}

	tagID, err := mgr.CreateTag(ctx, &tags.Tag{
		Name:       tagName,
		CategoryID: catID,
	})
	return tagID, nil
}

func isTagAttached(t *testing.T, ctx context.Context, mgr *tags.Manager, ref mo.Reference, tagID string) (bool, error) {
	t.Helper()

	attachedTags, err := mgr.GetAttachedTags(ctx, ref)
	if err != nil {
		return false, err
	}

	for _, tag := range attachedTags {
		if tag.ID == tagID {
			return true, nil
		}
	}

	return false, nil
}
