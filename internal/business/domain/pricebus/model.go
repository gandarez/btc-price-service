package pricebus

import (
	"time"

	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
)

// Price represents the price data structure.
type Price struct {
	Symbol    string
	Timestamp string
	Price     float64
}

func toBusPrice(asset coindeskclient.Asset) Price {
	ts := time.Unix(asset.PriceLastUpdatedAt, 0).UTC()

	return Price{
		Symbol:    asset.Symbol,
		Timestamp: ts.Format(time.RFC3339),
		Price:     asset.Price,
	}
}
