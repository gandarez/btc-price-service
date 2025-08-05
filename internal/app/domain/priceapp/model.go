package priceapp

import (
	"time"

	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
)

// Price represents the price data structure.
type Price struct {
	Symbol    string  `json:"symbol"`
	UpdatedAt string  `json:"timestamp"`
	Price     float64 `json:"price"`
}

// Timestamp returns the time when the price was last updated.
// It parses the UpdatedAt field which is expected to be in RFC3339 format.
func (p Price) Timestamp() time.Time {
	// it's safe to assume that UpdatedAt is in RFC3339 format
	// since it was formatted that way in the business layer.
	t, _ := time.Parse(time.RFC3339, p.UpdatedAt)
	return t
}

func toAppPrice(busPrice pricebus.Price) Price {
	return Price{
		Symbol:    busPrice.Symbol,
		UpdatedAt: busPrice.Timestamp,
		Price:     busPrice.Price,
	}
}
