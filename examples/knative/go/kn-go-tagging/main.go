package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/signals"

	"github.com/vmware-samples/vcenter-event-broker-appliance/examples/knative/go/kn-go-tagging/tagging"
)

func main() {
	var envConfig tagging.EnvConfig
	err := envconfig.Process("", &envConfig)
	if err != nil {
		panic(err.Error())
	}

	ctx := signals.NewContext()

	var logger *zap.Logger
	if envConfig.DebugLogs {
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err.Error())
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			panic(err.Error())
		}
	}

	if err := validateEnvConfig(envConfig); err != nil {
		logger.Fatal("configuration validation", zap.Error(err))
	}

	ctx = logging.WithLogger(ctx, logger.Sugar())
	c, err := tagging.NewClient(ctx)
	if err != nil {
		logger.Fatal("create client", zap.Error(err))
	}

	if err = c.Run(ctx); err != nil {
		logger.Fatal("run client receiver", zap.Error(err))
	}
}

func validateEnvConfig(config tagging.EnvConfig) error {
	if config.TagAction != "attach" && config.TagAction != "detach" {
		return fmt.Errorf("unrecognized tag action %q", config.TagAction)
	}
	return nil
}
