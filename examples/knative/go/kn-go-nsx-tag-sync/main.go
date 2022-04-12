package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/embano1/vsphere/logger"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vmware-samples/vcenter-event-broker-appliance/examples/knative/go/kn-go-nsx-tag-sync/tags"
)

var (
	buildCommit = "unknown"
	buildTag    = "unknown"
)

func main() {
	var cfg tags.Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic("process environment variables: " + err.Error())
	}

	log, err := getLogger(cfg.Debug)
	if err != nil {
		panic("create logger: " + err.Error())
	}
	log = log.Named("nsx-tag-sync")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx = logger.Set(ctx, log)

	syncer, err := tags.NewSyncer(ctx)
	if err != nil {
		log.Fatal("could not create vsphere to nsx tag synchronizer", zap.Error(err))
	}

	log.Info("starting vsphere to nsx tag synchronizer",
		zap.Int("listenPort", cfg.Port),
		zap.Bool("debug", cfg.Debug),
	)

	if err = syncer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal("could not run vsphere to nsx tag synchronizer", zap.Error(err))
	}
	log.Info("shutdown complete")
}

func getLogger(debug bool) (*zap.Logger, error) {
	fields := []zap.Field{
		zap.String("commit", buildCommit),
		zap.String("tag", buildTag),
	}

	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	}

	log, err := config.Build(zap.Fields(fields...))
	if err != nil {
		return nil, err
	}

	return log, nil
}
