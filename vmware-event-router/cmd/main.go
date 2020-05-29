package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/color"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/connection"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/stream"
	"golang.org/x/sync/errgroup"
)

var (
	commit  = "UNKNOWN"
	version = "UNKNOWN"
)

var banner = `
 _    ____  ___                            ______                 __     ____              __           
| |  / /  |/  /      ______ _________     / ____/   _____  ____  / /_   / __ \____  __  __/ /____  _____
| | / / /|_/ / | /| / / __  / ___/ _ \   / __/ | | / / _ \/ __ \/ __/  / /_/ / __ \/ / / / __/ _ \/ ___/
| |/ / /  / /| |/ |/ / /_/ / /  /  __/  / /___ | |/ /  __/ / / / /_   / _, _/ /_/ / /_/ / /_/  __/ /    
|___/_/  /_/ |__/|__/\__,_/_/   \___/  /_____/ |___/\___/_/ /_/\__/  /_/ |_|\____/\__,_/\__/\___/_/     

`

func main() {
	fmt.Println(banner)
	var logger = log.New(os.Stdout, color.Green("[VMware Event Router] "), log.LstdFlags)

	var configPath string
	var verbose bool
	var err error

	flag.StringVar(&configPath, "config", "/etc/vmware-event-router/config", "path to configuration file for metrics, stream source and processor")
	flag.BoolVar(&verbose, "verbose", false, "print event handling information")
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
		time.Sleep(3 * time.Second)
	}()

	f, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("could not open configuration file: %v", err)
	}
	cfgs, err := connection.Parse(f)
	if err != nil {
		logger.Fatalf("could not parse configuration file: %v", err)
	}

	var (
		streamer      stream.Streamer
		proc          processor.Processor
		metricsServer *metrics.Server // will be set if valid configuration provided
		bindAddr      string
	)

	// TODO: support multiple streams/processors. Current behavior: if multiple
	// definitions of the same type are given, last wins.
	for _, cfg := range cfgs {
		switch cfg.Type {
		case "stream":
			switch cfg.Provider {
			case stream.ProviderVSphere:
				logger.Printf("connecting to vCenter %s", cfg.Address)
				streamer, err = stream.NewVCenterStream(ctx, cfg, metricsServer, stream.WithVCenterVerbose(verbose))
				if err != nil {
					logger.Fatalf("could not connect to vCenter: %v", err)
				}

			default:
				logger.Fatalf("unsupported stream provider: %s", cfg.Provider)
			}

		case "processor":
			switch cfg.Provider {
			case processor.ProviderOpenFaaS:
				var async bool
				if cfg.Options["async"] == "true" {
					async = true
				}
				logger.Printf("connecting to OpenFaaS gateway %s (async mode: %v)", cfg.Address, async)
				proc, err = processor.NewOpenFaaSProcessor(ctx, cfg, metricsServer, processor.WithOpenFaaSVerbose(verbose))
				if err != nil {
					logger.Fatalf("could not connect to OpenFaaS: %v", err)
				}
			case processor.ProviderAWS:
				logger.Printf("connecting to AWS EventBridge (arn: %s)", cfg.Options["aws_eventbridge_rule_arn"])
				proc, err = processor.NewAWSEventBridgeProcessor(ctx, cfg, metricsServer, processor.WithAWSVerbose(verbose))
				if err != nil {
					logger.Fatalf("could not connect to AWS EventBridge: %v", err)
				}
			default:
				logger.Fatalf("unsupported processor provider: %s", cfg.Provider)
			}

		case "metrics":
			metricsServer, err = metrics.NewServer(cfg)
			bindAddr = cfg.Address
			if err != nil {
				logger.Fatalf("could not initialize metrics server: %v", err)
			}
			logger.Printf("exposing metrics server on %s (auth: %s)", cfg.Address, cfg.Auth.Method)

		default:
			logger.Fatalf("invalid type specified: %s", cfg.Type)
		}
	}

	// validate if the configuration provided is complete
	switch {
	case streamer == nil:
		logger.Fatal("no configuration for event stream provider found")
	case proc == nil:
		logger.Fatal("no configuration for event processor found")
	case metricsServer == nil:
		logger.Fatal("no configuration for metrics server found")
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// metrics server
	eg.Go(func() error {
		return metricsServer.Run(egCtx, bindAddr)
	})

	// event stream
	eg.Go(func() error {
		defer func() {
			_ = streamer.Shutdown(egCtx)
		}()
		return streamer.Stream(egCtx, proc)
	})

	// blocks
	err = eg.Wait()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("shutdown successful")
}
