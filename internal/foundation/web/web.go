package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gandarez/btc-price-service/internal/foundation/log"
)

type ResponseStream struct {
	Msg string
	Err error
}

// Encoder defines behavior that can encode a data model and provide
// the content type for that encoding.
type Encoder interface {
	Encode() (data []byte, contentType string, err error)
}

// HandlerFunc defines a function type for handling HTTP requests.
type HandlerFunc func(ctx context.Context, r *http.Request) Encoder

// HandlerFunc defines a function type for handling HTTP requests.
// type HandlerFuncStream func(ctx context.Context, r *http.Request, out chan ResponseStream)
type HandlerFuncStream func(w http.ResponseWriter, r *http.Request)

// App is the main application struct that holds the HTTP mux.
type App struct {
	mux *http.ServeMux
}

// NewApp creates a new App instance with an initialized HTTP mux.
func NewApp() *App {
	mux := http.NewServeMux()

	return &App{
		mux: mux,
	}
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// HandlerFunc registers a handler function for a specific HTTP method and path.
func (a *App) HandlerFunc(ctx context.Context, method, group, path string, handlerFunc HandlerFunc) {
	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	h := func(w http.ResponseWriter, r *http.Request) {
		rctx := r.Context()

		resp := handlerFunc(rctx, r)

		if err := respond(rctx, w, resp); err != nil {
			logger := log.Extract(ctx)
			logger.Errorf("Error processing request for %s: %s", finalPath, err)

			return
		}
	}

	a.mux.HandleFunc(finalPath, h)
}

// HandlerFuncStream registers a handler function for streaming responses.
func (a *App) HandlerFuncStream(
	ctx context.Context,
	group, path string,
	handlerFunc HandlerFuncStream,
) {
	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(w, r.WithContext(ctx))
	}

	a.mux.HandleFunc(finalPath, h)
}
