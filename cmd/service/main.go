package main

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gandarez/btc-price-service/internal/foundation/log"
	"github.com/gandarez/btc-price-service/internal/foundation/version"
)

const (
	// serviceName is the name of the service.
	serviceName = "btc-price-service"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	logger := log.New(os.Stdout)
	logger.WithFields([]zapcore.Field{
		zap.String("service", serviceName),
		zap.String("version", version.Version),
		zap.String("environment", "development"),
	}...)

	// Save logger to context
	ctx = log.ToContext(ctx, logger)

	logger.Infof("Service %s is starting", serviceName)
}
