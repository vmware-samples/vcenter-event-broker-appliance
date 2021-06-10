package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"knative.dev/pkg/signals"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor/aws"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor/knative"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/processor/openfaas"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/vcenter"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/provider/vcsim"
)

var (
	commit  = "UNKNOWN"
	version = "UNKNOWN"
)

const (
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
		configPath string
		logLevel   string
		logJSON    bool
	)

	flag.StringVar(&configPath, "config", defaultConfigPath, "path to configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "set log level (debug,info,warn,error)")
	flag.BoolVar(&logJSON, "log-json", false, "print JSON-formatted logs")
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("\ncommit: %s\n", commit)
		fmt.Printf("version: %s\n", version)
	}
	flag.Parse()

	var lvl zapcore.Level
	err := lvl.Set(logLevel)
	if err != nil {
		panic(err.Error())
	}

	// configure logger using defaults from the zap prod config
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.Level = zap.NewAtomicLevelAt(lvl)
	if !logJSON {
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := zapCfg.Build(zap.AddStacktrace(zap.ErrorLevel)) // stack traces only error and above
	if err != nil {
		panic(err.Error())
	}
	log := logger.Named("[MAIN]").Sugar().With("commit", commit, "version", version)

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("could not open configuration file: %v", err)
	}

	cfg, err := config.Parse(f)
	if err != nil {
		log.Fatalf("could not parse configuration file: %v", err)
	}

	var (
		prov provider.Provider
		proc processor.Processor
		ms   *metrics.Server // allows nil check
	)

	ctx := signals.NewContext()

	// set up event provider
	switch cfg.EventProvider.Type {
	case config.ProviderVCenter:
		prov, err = vcenter.NewEventStream(ctx, cfg.EventProvider.VCenter, ms, logger.Sugar(), vcenter.WithRootCAs(cfg.Certificates.RootCAs))
		if err != nil {
			log.Fatalf("could not connect to vCenter: %v", err)
		}

		log.Infow("connecting to vCenter", "address", cfg.EventProvider.VCenter.Address)

	case config.ProviderVCSIM:
		log.Warn("%s is DEPRECATED and will be removed in future versions", config.ProviderVCSIM)
		prov, err = vcsim.NewEventStream(ctx, cfg.EventProvider.VCSIM, ms, logger.Sugar())
		if err != nil {
			log.Fatalf("could not connect to vCenter simulator: %v", err)
		}

		log.Infow("connecting to vCenter simulator", "address", cfg.EventProvider.VCSIM.Address)

	default:
		log.Fatalf("invalid type specified: %q", cfg.EventProvider.Type)
	}

	// set up event processor
	switch cfg.EventProcessor.Type {
	case config.ProcessorOpenFaaS:
		proc, err = openfaas.NewProcessor(ctx, cfg.EventProcessor.OpenFaaS, ms, logger.Sugar())
		if err != nil {
			log.Fatalf("could not connect to OpenFaaS: %v", err)
		}

		log.Infow("connected to OpenFaaS gateway", "address", cfg.EventProcessor.OpenFaaS.Address, "async", cfg.EventProcessor.OpenFaaS.Async)

	case config.ProcessorEventBridge:
		proc, err = aws.NewEventBridgeProcessor(ctx, cfg.EventProcessor.EventBridge, ms, logger.Sugar())
		if err != nil {
			log.Fatalf("could not connect to AWS EventBridge: %v", err)
		}

		log.Infow("connected to AWS EventBridge", "ruleARN", cfg.EventProcessor.EventBridge.RuleARN)

	case config.ProcessorKnative:
		proc, err = knative.NewProcessor(ctx, cfg.EventProcessor.Knative, ms, logger.Sugar())
		if err != nil {
			log.Fatalf("could not create Knative processor: %v", err)
		}

		log.Infow("created Knative processor", "sink", proc.(*knative.Processor).Sink())

	default:
		log.Fatalf("invalid type specified: %q", cfg.EventProcessor.Type)
	}

	// set up metrics provider (only supporting default for now)
	switch cfg.MetricsProvider.Type {
	case config.MetricsProviderDefault:
		ms, err = metrics.NewServer(cfg.MetricsProvider.Default, logger.Sugar())
		if err != nil {
			log.Fatalf("could not initialize metrics server: %v", err)
		}

	default:
		log.Fatalf("invalid type specified: %q", cfg.MetricsProvider.Type)
	}

	// validate if the configuration provided is complete
	switch {
	case prov == nil:
		log.Fatal("no valid configuration for event provider found")
	case proc == nil:
		log.Fatal("no valid configuration for event processor found")
	case ms == nil:
		log.Fatal("no valid configuration for metrics server found")
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// metrics server
	eg.Go(func() error {
		return ms.Run(egCtx)
	})

	// event stream
	eg.Go(func() error {
		return prov.Stream(egCtx, proc)
	})

	// shutdown handling
	eg.Go(func() error {
		<-egCtx.Done()
		log.Infof("initiating shutdown")

		var shutdownErr []error
		err = prov.Shutdown(egCtx)
		if err != nil {
			shutdownErr = append(shutdownErr, fmt.Errorf("could not gracefully shutdown provider: %v", err))
		}

		err = proc.Shutdown(egCtx)
		if err != nil {
			shutdownErr = append(shutdownErr, fmt.Errorf("could not gracefully shutdown processor: %v", err))
		}

		if shutdownErr == nil {
			log.Info("shutdown successful")
			return nil
		}

		for i, sdErr := range shutdownErr {
			log.Warnf("shutdown error [%d]: %v", i, sdErr)
		}
		return nil // don't propagate shutdown errors
	})

	// blocks
	err = eg.Wait()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Fatal(err)
		}
	}
}
