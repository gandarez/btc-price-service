package priceapp

import (
	"testing"
	"time"

	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
	"github.com/stretchr/testify/assert"
)

func TestPrice_Encode(t *testing.T) {
	price := pricebus.Price{
		Symbol:    "BTC",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Price:     50000.0,
	}

	p := toAppPrice(price)

	assert.Equal(t, p.Symbol, price.Symbol)
	assert.Equal(t, p.UpdatedAt, price.Timestamp)
	assert.Equal(t, p.Price, price.Price)
}
