package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gandarez/btc-price-service/internal/app/domain/checkapp"
	"github.com/gandarez/btc-price-service/internal/app/domain/priceapp"
	"github.com/gandarez/btc-price-service/internal/app/sdk/mux"
	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
	"github.com/gandarez/btc-price-service/internal/foundation/config"
	"github.com/gandarez/btc-price-service/internal/foundation/log"
	"github.com/gandarez/btc-price-service/internal/foundation/version"
	"github.com/gandarez/btc-price-service/internal/foundation/web"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	logger := log.New(os.Stdout)

	cfgPath, err := filepath.Abs("./configs/.env")
	if err != nil {
		logger.Fatalf("failed to get absolute path for env file: %s", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Fatalf("failed to load config: %v", err)
	}

	logger.WithFields([]zapcore.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("version", version.Version),
		zap.String("environment", cfg.Environment),
	}...)

	// Save logger to context
	ctx = log.ToContext(ctx, logger)

	logger.Infof("service %s is starting..", cfg.ServiceName)
	logger.Infof("params: %s", cfg.String())

	// Initialize price bus
	coindeskcli := coindeskclient.NewClient(
		cfg.CoinDeskConfig.URL,
		cfg.CoinDeskConfig.APIKey,
	)
	priceBus := pricebus.NewBusiness(coindeskcli)

	// build http routes
	cfgMux := mux.Config{
		PriceConfig: mux.PriceConfig{
			BufferTTL:                 time.Duration(cfg.CacheConfig.TTL) * time.Second,
			MaxCacheSize:              cfg.CacheConfig.MaxSize,
			DefaultExpirationInterval: time.Duration(cfg.CacheConfig.ExpirationInterval) * time.Second,
			PollInterval:              time.Duration(cfg.CoinDeskConfig.PollInterval) * time.Second,
			MaxPeersPerBroadcaster:    cfg.BroadcastConfig.MaxPeersPerBroadcaster,
			PriceBus:                  priceBus,
		},
	}

	mux := mux.WebAPI(ctx, cfgMux, buildRoutes())
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerConfig.Port),
		ReadHeaderTimeout: time.Duration(cfg.ServerConfig.ReadHeaderTimeout) * time.Second,
		Handler:           mux,
	}

	// Start http server
	serverError := make(chan error, 1)

	go func() {
		logger.Infof("http server started on %s", server.Addr)

		serverError <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverError:
		logger.Fatalf("server error: %v", err)
	case <-shutdown:
		logger.Infoln("received shutdown signal, shutting down service")

		ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimeout)*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("failed to shutdown server: %v", err)
		}

		logger.Infof("service %s has shut down", cfg.ServiceName)
	}

	logger.Infof("service %s gracefully stopped", cfg.ServiceName)
}

type add struct{}

func buildRoutes() mux.RouteAdder {
	return add{}
}

func (add) Add(ctx context.Context, app *web.App, cfg mux.Config) {
	checkapp.Routes(ctx, app)
	priceapp.Routes(ctx, app, cfg)
}
