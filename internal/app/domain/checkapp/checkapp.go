package checkapp

import (
	"context"
	"net/http"
	"os"

	"github.com/gandarez/btc-price-service/internal/foundation/version"
	"github.com/gandarez/btc-price-service/internal/foundation/web"
)

type app struct {
}

func newApp() *app {
	return &app{}
}

func (*app) readiness(_ context.Context, _ *http.Request) web.Encoder {
	return nil
}

func (*app) liveness(_ context.Context, _ *http.Request) web.Encoder {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	return Info{
		Status:   "OK",
		Version:  version.Version,
		Hostname: host,
	}
}
