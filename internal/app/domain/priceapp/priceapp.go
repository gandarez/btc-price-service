package priceapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gandarez/btc-price-service/internal/app/sdk/pubsub"
	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
	"github.com/gandarez/btc-price-service/internal/foundation/cache"
	"github.com/gandarez/btc-price-service/internal/foundation/log"
)

type (
	app struct {
		priceBus    PriceBusiness
		broadcaster *pubsub.Manager
		cache       *cache.Buffer[cache.CacheableEntity]
		cfg         Config
	}

	// Config holds the configuration for the price application.
	Config struct {
		BufferTTL                 time.Duration
		MaxCacheSize              int
		DefaultExpirationInterval time.Duration
		PollInterval              time.Duration
		MaxPeersPerBroadcaster    int
		PriceBus                  *pricebus.Business
	}

	// PriceBusiness defines the interface for fetching asset prices.
	PriceBusiness interface {
		AssetPrice(ctx context.Context, symbol string, pagination page.Page) (pricebus.Price, error)
	}
)

func newApp(cfg Config) *app {
	return &app{
		priceBus:    cfg.PriceBus,
		broadcaster: pubsub.NewManager(cfg.MaxPeersPerBroadcaster),
		cache:       cache.NewBuffer[cache.CacheableEntity](cfg.BufferTTL, cfg.DefaultExpirationInterval, cfg.MaxCacheSize),
		cfg:         cfg,
	}
}

func (a *app) startPolling(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	logger := log.Extract(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			page, err := page.New(1, 100) // Default to page 1 with 100 rows per page
			if err != nil {
				logger.Errorf("failed to create pagination: %v", err)
				continue
			}

			price, err := a.priceBus.AssetPrice(ctx, "BTC", page)
			if err != nil {
				logger.Errorf("failed to fetch asset price: %v", err)
				continue
			}

			update := toAppPrice(price)

			// if cached item is equal to current, then do not broadcast
			if last, ok := a.cache.Last().(Price); ok && last.Price == update.Price {
				logger.Infof("skipping broadcast for unchanged price: %v", update)
				continue
			}

			logger.Infof("broadcasting update: %v", update)

			a.cache.Add(update) // cache for reconnection if needed
			a.broadcaster.Broadcast(update)
		}
	}
}

func (a *app) priceStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	logger := log.Extract(ctx)

	// Send initial message
	_, err := fmt.Fprintln(w, ": connected")
	if err != nil {
		logger.Errorf("failed to send initial message: %s", err)
	}

	flusher.Flush()

	since, err := a.parsePriceStreamParams(r)
	if err != nil {
		logger.Errorf("failed to parse price-stream params: %s", err)

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	sub := a.broadcaster.Subscribe(ctx)
	defer a.broadcaster.Unsubscribe(ctx, sub)

	if !since.IsZero() {
		wait := make(chan struct{}, 1)

		logger.Infof("fetching prices since: %s", since)

		go func() {
			defer close(wait)
			// Stream missed updates
			for _, update := range a.cache.Since(since) {
				if err := sendSSE(w, update); err != nil {
					logger.Infof("client disconnected from price stream (send failed): %s", err)

					return
				}

				flusher.Flush()
			}
		}()

		<-wait
	}

	logger.Infoln("client connected to price stream")

	// send last price if available
	if last, ok := a.cache.Last().(Price); ok {
		a.broadcaster.SendOne(sub, last)
	}

	// add periodic ping to detect disconnections
	pingTicker := time.NewTicker(2 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Infoln("client disconnected from price stream")

			return
		case <-pingTicker.C:
			// Send ping to detect if client is still connected
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				logger.Infof("client disconnected from price stream (ping failed): %s", err)

				return
			}

			flusher.Flush()
		case update := <-sub.Ch:
			if err := sendSSE(w, update); err != nil {
				logger.Infof("client disconnected from price stream (send failed): %s", err)

				return
			}

			flusher.Flush()
		}
	}
}

// sendSSE sends a Server-Sent Event (SSE) to the client.
func sendSSE(w http.ResponseWriter, update cache.CacheableEntity) error {
	data, _ := json.Marshal(update)
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)

	return err
}

func (a *app) parsePriceStreamParams(r *http.Request) (time.Time, error) {
	sinceStr := r.URL.Query().Get("since")

	if sinceStr == "" {
		return time.Time{}, nil // No 'since' parameter, return zero time
	}

	sinceTime, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid 'since' timestamp format: %v", err)
	}

	logger := log.Extract(r.Context())
	logger.Infof("Parsed 'since' timestamp: %s", sinceStr)
	logger.Infof("time now: %s", time.Now().UTC().Format(time.RFC3339))

	cutoff := time.Now().UTC().Add(-a.cfg.BufferTTL)
	if sinceTime.Before(cutoff) {
		return time.Time{}, errors.New("'since' timestamp is too old (exceeds buffer TTL)")
	}

	return sinceTime, nil
}
