package priceapp

import (
	"context"

	"github.com/gandarez/btc-price-service/internal/app/sdk/mux"
	"github.com/gandarez/btc-price-service/internal/foundation/web"
)

// Routes registers the routes for the price application.
func Routes(ctx context.Context, app *web.App, cfg mux.Config) {
	const version = "v1"

	api := newApp(Config{
		BufferTTL:                 cfg.PriceConfig.BufferTTL,
		MaxCacheSize:              cfg.PriceConfig.MaxCacheSize,
		DefaultExpirationInterval: cfg.PriceConfig.DefaultExpirationInterval,
		PollInterval:              cfg.PriceConfig.PollInterval,
		PriceBus:                  cfg.PriceConfig.PriceBus,
	})

	go api.startPolling(ctx)

	app.HandlerFuncStream(ctx, version, "/price-stream", api.priceStream)
}
