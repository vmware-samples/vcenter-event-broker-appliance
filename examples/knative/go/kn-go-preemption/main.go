package main

import (
	"context"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

var (
	buildCommit = "undefined" // build injection
	buildTag    = "undefined" // build injection
)

func main() {
	ctx := context.Background()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		panic("unable to parse environment variables: " + err.Error())
	}

	var logger *zap.Logger
	if env.Debug {
		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			panic("unable to create logger: " + err.Error())
		}
		logger = zapLogger

	} else {
		zapLogger, err := zap.NewProduction()
		if err != nil {
			panic("unable to create logger: " + err.Error())
		}
		logger = zapLogger
	}

	logger = logger.With(zap.String("commit", buildCommit), zap.String("tag", buildTag))
	ctx = logging.WithLogger(ctx, logger.Sugar())
	c, err := newClient(ctx, logger)
	if err != nil {
		logger.Sugar().Fatalw("could not create temporal client", zap.Error(err))
	}
	defer c.close()

	ceClient, err := ce.NewClientHTTP(http.WithPort(env.Port))
	if err != nil {
		logger.Sugar().Fatalw("could not create http client", zap.Error(err))
	}

	logger.Sugar().Infow("starting listener", "port", env.Port)
	if err = ceClient.StartReceiver(ctx, c.handler); err != nil {
		logger.Sugar().Fatalw("could not start alarm event handler", zap.Error(err))
	}
}
