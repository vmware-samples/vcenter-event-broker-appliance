package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	handler "github.com/openfaas/templates-sdk/go-http"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const cfgPath = "/var/openfaas/secrets/vcconfig"

var (
	vsClt *vsClient // Persist vSphere connection.
	once  sync.Once // For handleSignal() to be called once.
	debug bool      // True for verbose logging.
)

type vmConfig struct {
	name    string
	valBool bool
	valInt  int
}

// Handle a function invocation
func Handle(req handler.Request) (handler.Response, error) {
	ctx := context.Background()
	debug = setVerbosity()

	// Load config every time, to ensure the most updated version is used.
	vcCfg, err := loadConfig(cfgPath)
	if err != nil {
		return handlerResponseWithError("load vcconfig: %w",
			http.StatusBadRequest,
			err,
		)
	}

	// The Mananged Object Reference for the VM that powered on.
	vmMOR, err := eventVmMOR(req.Body)
	if err != nil {
		return handlerResponseWithError(
			"retrieve VM object: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	if vsClt == nil {
		// Save vSphere connection in a global variable.
		vsClt, err = newClient(ctx, vcCfg)
		if err != nil {
			return handlerResponseWithError(
				"connect to vSphere: %w",
				http.StatusUnauthorized,
				err,
			)
		}
	}

	// Signals wil cause vSphere APIs to log out.
	once.Do(func() {
		go handleSignal(ctx, vsClt)
	})

	tagObjs, err := vsClt.attachedTags(ctx, *vmMOR)
	if err != nil {
		// Try to reconnect to vSphere if there's an auth error.
		err = authErrReconnect(ctx, vcCfg, err, "401 Unauthorized")
		if err == nil {
			tagObjs, err = vsClt.attachedTags(ctx, *vmMOR)
		}
	}

	if err != nil {
		return handlerResponseWithError(
			"get tags attached to VM: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	// Check SOAP connection by getting managed object VM.
	moVM, err := vsClt.moVirtualMachine(ctx, *vmMOR)
	if err != nil {
		// Try to reconnect to vSphere if there's an auth error.
		err = authErrReconnect(ctx, vcCfg, err, "NotAuthenticated")
		if err == nil {
			moVM, err = vsClt.moVirtualMachine(ctx, *vmMOR)
		}
	}

	if err != nil {
		return handlerResponseWithError(
			"retrieve current VM configs: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	// Get the current configurations in a list from the VM object.
	currCfgs := currentConfigs(ctx, moVM)

	// Look for configurations that need to be applied to the VM.
	newCfgs, err := vsClt.unappliedConfigs(ctx, tagObjs, currCfgs)
	if err != nil {
		return handlerResponseWithError(
			"determine unapplied configs: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	var task *object.Task

	if len(newCfgs) > 0 {
		task, err = vsClt.applyCfgs(ctx, moVM, *vmMOR, newCfgs)
		if err != nil {
			return handlerResponseWithError(
				"apply config tags: %w",
				http.StatusInternalServerError,
				err,
			)
		}
	}

	message, err := appliedConfigsMessage(task, newCfgs)
	if err != nil {
		return handlerResponseWithError(
			"create response message: %w",
			http.StatusInternalServerError,
			err,
		)
	}

	fmt.Printf("Success (%v): %s.\n", http.StatusAccepted, message)

	return handler.Response{
		Body:       message,
		StatusCode: http.StatusAccepted,
	}, nil
}

func (c *vsClient) attachedTags(ctx context.Context, mor types.ManagedObjectReference) ([]tags.Tag, error) {
	// Tag objects attached to the VM.
	tagObjs, err := c.tagMgr.GetAttachedTags(ctx, mor)
	if err != nil {
		return []tags.Tag{}, fmt.Errorf("get attached tags: %w", err)
	}

	if debug {
		log.Println("Retrieved", len(tagObjs), "tag(s).")
	}

	return tagObjs, nil
}

func (c *vsClient) moVirtualMachine(ctx context.Context, mor types.ManagedObjectReference) (mo.VirtualMachine, error) {
	var moVM mo.VirtualMachine
	pc := property.DefaultCollector(c.govmomi.Client)

	err := pc.RetrieveOne(ctx, mor, []string{}, &moVM)
	if err != nil {
		return mo.VirtualMachine{}, fmt.Errorf("retrieve VM configurations :%w", err)
	}

	if debug {
		log.Println("Retrieve VM configs", moVM)
	}

	return moVM, nil
}

// currentConfigs gets a list of current selected configs set for the VM object.
func currentConfigs(ctx context.Context, moVM mo.VirtualMachine) []vmConfig {
	configs := make([]vmConfig, 6)

	configs[0] = vmConfig{name: "numCPU", valInt: int(moVM.Config.Hardware.NumCPU)}
	configs[1] = vmConfig{name: "memoryMB", valInt: int(moVM.Config.Hardware.MemoryMB)}
	configs[2] = vmConfig{name: "numCoresPerSocket", valInt: int(moVM.Config.Hardware.NumCoresPerSocket)}

	if moVM.Config.MemoryHotAddEnabled != nil {
		configs[3] = vmConfig{name: "memoryHotAddEnabled", valBool: *moVM.Config.MemoryHotAddEnabled}
	}

	if moVM.Config.CpuHotRemoveEnabled != nil {
		configs[4] = vmConfig{name: "cpuHotRemoveEnabled", valBool: *moVM.Config.CpuHotRemoveEnabled}
	}

	if moVM.Config.CpuHotAddEnabled != nil {
		configs[5] = vmConfig{name: "cpuHotRemoveEnabled", valBool: *moVM.Config.CpuHotAddEnabled}
	}

	if debug {
		log.Println("Current VM configurations:", configs)
	}

	return configs
}

// selectConfigs determines which tagged configurations and their values
// that need to be applied.
func (c *vsClient) unappliedConfigs(ctx context.Context, tagObjs []tags.Tag, currCfgs []vmConfig) ([]vmConfig, error) {
	// Get config information from tags.
	tagCfgs, err := c.desiredConfigs(ctx, tagObjs)
	if err != nil {
		return []vmConfig{}, fmt.Errorf("get config info from tags: %w", err)
	}

	unappliedCfgs := filterOutCurrentConfigs(tagCfgs, currCfgs)

	if debug {
		log.Println("List of unapplied configs:", unappliedCfgs)
	}

	return unappliedCfgs, nil
}

// Figure out what the desired VM configs are.
func (c *vsClient) desiredConfigs(ctx context.Context, attachedTags []tags.Tag) ([]vmConfig, error) {
	// These are the category names that represent VM configurations we can change.
	wantedCategories := []string{
		"config.hardware.numCPU",
		"config.hardware.memoryMB",
		"config.hardware.numCoresPerSocket",
		"config.memoryHotAddEnabled",
		"config.cpuHotRemoveEnabled",
		"config.cpuHotAddEnabled",
	}

	cfgs := []vmConfig{}

	for _, attTag := range attachedTags {
		catObj, err := c.tagMgr.GetCategory(ctx, attTag.CategoryID)
		if err != nil {
			return []vmConfig{}, fmt.Errorf("get category from catID: %w", err)
		}

		for _, want := range wantedCategories {
			if catObj.Name == want {
				// Remove the 'config.hardware.' prefix.
				cfgType := strings.TrimPrefix(catObj.Name, "config.")
				cfgType = strings.TrimPrefix(cfgType, "hardware.")

				cfg, err := buildConfig(attTag, cfgType)
				if err != nil {
					return []vmConfig{}, fmt.Errorf("build config: %w", err)
				}

				cfgs = append(cfgs, cfg)
			}
		}
	}

	return cfgs, nil
}

// buildConfig will get the config value from an attached tag. Then, it will return
// a VM config containing the config type and value.
func buildConfig(attachedTag tags.Tag, cfgType string) (vmConfig, error) {
	if cfgType == "numCPU" || cfgType == "memoryMB" || cfgType == "numCoresPerSocket" {
		val, err := strconv.Atoi(attachedTag.Name)
		if err != nil {
			return vmConfig{}, fmt.Errorf("convert string to int :%w", err)
		}

		return vmConfig{name: cfgType, valInt: val}, nil
	}

	if cfgType == "memoryHotAddEnabled" || cfgType == "cpuHotRemoveEnabled" || cfgType == "cpuHotAddEnabled" {
		val, err := strconv.ParseBool(attachedTag.Name)
		if err != nil {
			return vmConfig{}, fmt.Errorf("convert string to bool: %w", err)
		}

		return vmConfig{name: cfgType, valBool: val}, nil
	}

	return vmConfig{}, errors.New(cfgType + " is not an expected config type")
}

// filterOutCurrentConfigs removes the tag configs that are already current configs.
func filterOutCurrentConfigs(tagConfigs, currConfigs []vmConfig) []vmConfig {
	unappliedCfgs := []vmConfig{}

	for _, tc := range tagConfigs {
		if !isCfgCurrent(tc, currConfigs) {
			unappliedCfgs = append(unappliedCfgs, tc)
		}
	}
	return unappliedCfgs
}

// isHwCfgMatch determines if the given tag's hardware configuration
// is already the current configuration of the hardware.
func isCfgCurrent(tagConfig vmConfig, currConfigs []vmConfig) bool {
	for _, curr := range currConfigs {
		if tagConfig == curr {
			return true
		}
	}

	return false
}

// makeCfgsMatch sets configuration of the VM to that of the attached tag.
func (c *vsClient) applyCfgs(ctx context.Context, moVM mo.VirtualMachine, mor types.ManagedObjectReference, cfgs []vmConfig) (*object.Task, error) {
	vm := object.NewVirtualMachine(c.govmomi.Client, mor)
	desiredSpec := generateDesiredSpec(cfgs, moVM)

	task, err := vm.Reconfigure(ctx, desiredSpec)
	if err != nil {
		return nil, err
	}

	if debug {
		log.Println("VM reconfigured with", desiredSpec)
	}

	return task, nil
}

func generateDesiredSpec(cfgs []vmConfig, moVM mo.VirtualMachine) types.VirtualMachineConfigSpec {
	var spec types.VirtualMachineConfigSpec

	for _, c := range cfgs {
		switch c.name {
		case "numCPU":
			spec.NumCPUs = int32(c.valInt)
		case "memoryMB":
			spec.MemoryMB = int64(c.valInt)
		case "numCoresPerSocket":
			spec.NumCoresPerSocket = int32(c.valInt)
		case "memoryHotAddEnabled":
			spec.MemoryHotAddEnabled = &c.valBool
		case "cpuHotRemoveEnabled":
			spec.CpuHotRemoveEnabled = &c.valBool
		case "cpuHotAddEnabled":
			spec.CpuHotAddEnabled = &c.valBool
		}
	}

	// Set ChangeVersion to guard against updates that have happened between when
	// configInfo is read and when it is applied.
	spec.ChangeVersion = moVM.Config.ChangeVersion

	return spec
}

type message struct {
	TaskID string
	Action string
}

func appliedConfigsMessage(task *object.Task, cfgs []vmConfig) ([]byte, error) {
	msg := message{
		Action: "Nothing to configure.",
	}

	if task != nil {
		msg.TaskID = task.Reference().Value
		act := "Set "

		for _, c := range cfgs {
			if c.name == "memoryHotAddEnabled" || c.name == "cpuHotRemoveEnabled" || c.name == "cpuHotAddEnabled" {
				act += fmt.Sprintf("%s to %v, ", c.name, c.valBool)
			} else {
				act += fmt.Sprintf("%s to %v, ", c.name, c.valInt)
			}
		}

		msg.Action = strings.TrimRight(act, ", ") + "."
	}

	jbs, err := json.Marshal(msg)
	if err != nil {
		return []byte{}, err
	}

	return jbs, nil
}
