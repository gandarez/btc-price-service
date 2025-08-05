package checkapp

import (
	"context"
	"net/http"

	"github.com/gandarez/btc-price-service/internal/foundation/web"
)

// Routes sets up the HTTP routes for the check/healthcheck application.
func Routes(ctx context.Context, app *web.App) {
	const version = "v1"

	api := newApp()

	app.HandlerFunc(ctx, http.MethodGet, version, "/readiness", api.readiness)
	app.HandlerFunc(ctx, http.MethodGet, version, "/liveness", api.liveness)
}
