package pricebus

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
)

func TestToBusPrice(t *testing.T) {
	asset := coindeskclient.Asset{
		Symbol:             "BTC",
		Price:              50000.0,
		PriceLastUpdatedAt: 1633072800,
	}

	price := toBusPrice(asset)

	assert.Equal(t, Price{
		Symbol:    "BTC",
		Price:     50000.0,
		Timestamp: "2021-10-01T07:20:00Z",
	}, price)
}
