package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/embano1/vsphere/logger"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	buildCommit = "unknown"
	buildTag    = "unknown"
)

type config struct {
	// http settings
	Address string `envconfig:"ADDRESS" default:"0.0.0.0" required:"true"`
	Path    string `envconfig:"WEBHOOK_PATH" default:"/webhook" required:"true"`

	// knative injected
	Port    int    `envconfig:"PORT" default:"8080" required:"true"`
	Service string `envconfig:"K_SERVICE" required:"true"`
	Sink    string `envconfig:"K_SINK" required:"true"`

	Debug bool `envconfig:"DEBUG" default:"false"`

	SecretPath string `envconfig:"WEBHOOK_SECRET_PATH"`
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		panic("process environment variables: " + err.Error())
	}

	log, err := getLogger(cfg.Debug)
	if err != nil {
		panic("create logger: " + err.Error())
	}
	log = log.Named("harbor-webhook")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx = logger.Set(ctx, log)

	if err = run(ctx, cfg); err != nil {
		log.Panic("could not run server", zap.Error(err))
	}

	log.Info("graceful shutdown complete")
}

func getLogger(debug bool) (*zap.Logger, error) {
	fields := []zap.Field{
		zap.String("commit", buildCommit),
		zap.String("tag", buildTag),
	}

	var cfg zap.Config
	if debug {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	}

	log, err := cfg.Build(zap.Fields(fields...))
	if err != nil {
		return nil, err
	}

	return log, nil
}
