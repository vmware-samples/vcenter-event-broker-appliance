package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/vcenter"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/vcsim"
)

var (
	commit  = "UNKNOWN"
	version = "UNKNOWN"
)

const (
	graceDelay        = 3 // delay when shutdown initiated
	defaultConfigPath = "/etc/vmware-event-router/config"
)

var banner = `
 _    ____  ___                            ______                 __     ____              __           
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /    
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/     

`

func main() {
	fmt.Print(banner)

	var (
		logger     = log.New(os.Stdout, color.Green("[VMware Event Router] "), log.LstdFlags)
		configPath string
		verbose    bool
		err        error
	)

	flag.StringVar(&configPath, "config", defaultConfigPath, "path to configuration file")
	flag.BoolVar(&verbose, "verbose", false, "verbose log output (default false)")
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\ncommit: %s\n", commit)
		fmt.Printf("version: %s\n", version)
	}
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// signal handler
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

		sig := <-sigCh
		logger.Printf("got signal: %v, cleaning up...", sig)

		cancel()
		// give goroutines some grace time to clean up
		time.Sleep(graceDelay * time.Second)
	}()

	f, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("could not open configuration file: %v", err)
	}

	cfg, err := config.Parse(f)
	if err != nil {
		logger.Fatalf("could not parse configuration file: %v", err)
	}

	var (
		prov provider.Provider
		proc processor.Processor
		ms   *metrics.Server // allows nil check
	)

	// set up event provider
	switch cfg.EventProvider.Type {
	case config.ProviderVCenter:
		prov, err = vcenter.NewEventStream(ctx, cfg.EventProvider.VCenter, ms, vcenter.WithVerbose(verbose))
		if err != nil {
			logger.Fatalf("could not connect to vCenter: %v", err)
		}

		logger.Printf("connecting to vCenter %q", cfg.EventProvider.VCenter.Address)

	case config.ProviderVCSIM:
		prov, err = vcsim.NewEventStream(ctx, cfg.EventProvider.VCSIM, ms, vcsim.WithVerbose(verbose))
		if err != nil {
			logger.Fatalf("could not connect to vCenter simulator: %v", err)
		}

		logger.Printf("connecting to vCenter simulator %q", cfg.EventProvider.VCSIM.Address)

	// TODO: implement
	// case config.ProviderVCD:

	default:
		logger.Fatalf("invalid type specified: %q", cfg.EventProvider.Type)
	}

	// set up event processor
	switch cfg.EventProcessor.Type {
	case config.ProcessorOpenFaaS:
		proc, err = processor.NewOpenFaaSProcessor(ctx, cfg.EventProcessor.OpenFaaS, ms, processor.WithOpenFaaSVerbose(verbose))
		if err != nil {
			logger.Fatalf("could not connect to OpenFaaS: %v", err)
		}

		logger.Printf("connected to OpenFaaS gateway %q (async mode: %t)", cfg.EventProcessor.OpenFaaS.Address, cfg.EventProcessor.OpenFaaS.Async)

	case config.ProcessorEventBridge:
		proc, err = processor.NewEventBridgeProcessor(ctx, cfg.EventProcessor.EventBridge, ms, processor.WithAWSVerbose(verbose))
		if err != nil {
			logger.Fatalf("could not connect to AWS EventBridge: %v", err)
		}

		logger.Printf("connected to AWS EventBridge using rule ARN %q", cfg.EventProcessor.EventBridge.RuleARN)

	default:
		logger.Fatalf("invalid type specified: %q", cfg.EventProcessor.Type)
	}

	// set up metrics provider (only supporting default for now)
	switch cfg.MetricsProvider.Type {
	case config.MetricsProviderDefault:
		ms, err = metrics.NewServer(cfg.MetricsProvider.Default)
		if err != nil {
			logger.Fatalf("could not initialize metrics server: %v", err)
		}

		logger.Printf("exposing metrics server on %s", cfg.MetricsProvider.Default.BindAddress)

	default:
		logger.Fatalf("invalid type specified: %q", cfg.MetricsProvider.Type)
	}

	// validate if the configuration provided is complete
	switch {
	case prov == nil:
		logger.Fatal("no valid configuration for event provider found")
	case proc == nil:
		logger.Fatal("no valid configuration for event processor found")
	case ms == nil:
		logger.Fatal("no valid configuration for metrics server found")
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// metrics server
	eg.Go(func() error {
		return ms.Run(egCtx)
	})

	// event stream
	eg.Go(func() error {
		defer func() {
			_ = prov.Shutdown(egCtx)
		}()
		return prov.Stream(egCtx, proc)
	})

	// blocks
	err = eg.Wait()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Fatal(err)
		}
	}

	logger.Println("shutdown successful")
}
