package mux

import (
	"context"
	"net/http"

	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
	"github.com/gandarez/btc-price-service/internal/foundation/web"
)

type (
	// PriceConfig holds the configuration for the price domain.
	PriceConfig struct {
		PriceBus *pricebus.Business
	}

	// Config holds the configuration for the mux.
	Config struct {
		PriceConfig PriceConfig
	}
)

// RouteAdder is an interface for adding routes to the web application.
type RouteAdder interface {
	Add(ctx context.Context, app *web.App, cfg Config)
}

// WebAPI initializes the web application with the provided route adder.
// It returns an http.Handler that serves the web application.
func WebAPI(ctx context.Context, cfg Config, routeAdder RouteAdder) http.Handler {
	app := web.NewApp()

	routeAdder.Add(ctx, app, cfg)

	return app
}
