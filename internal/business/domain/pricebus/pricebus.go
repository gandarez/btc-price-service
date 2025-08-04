package pricebus

import (
	"context"
	"fmt"

	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
)

type (
	// Business represents the business logic for the price domain.
	Business struct {
		coindeskcli CoinDeskClient
	}

	// CoinDeskClient defines the interface for fetching top asset prices from the CoinDesk API.
	CoinDeskClient interface {
		TopList(context.Context, page.Page) (coindeskclient.Result, error)
	}
)

// NewBusiness creates a new instance of the Business struct.
func NewBusiness(coindeskcli CoinDeskClient) *Business {
	return &Business{
		coindeskcli: coindeskcli,
	}
}

// AssetPrice retrieves the current asset price from the CoinDesk API.
func (b *Business) AssetPrice(ctx context.Context, symbol string, pagination page.Page) (Price, error) {
	result, err := b.coindeskcli.TopList(ctx, pagination)
	if err != nil {
		return Price{}, fmt.Errorf("failed to fetch top list: %v", err)
	}

	if result.Error != "" {
		return Price{}, fmt.Errorf("failed to fetch top list: %s", result.Error)
	}

	if len(result.TopList.Data.Assets) == 0 {
		return Price{}, fmt.Errorf("no assets found for page %d", pagination.Number())
	}

	for _, asset := range result.TopList.Data.Assets {
		if asset.Symbol == symbol {
			return toBusPrice(asset), nil
		}
	}

	// recursively search through pages until we find the BTC asset
	if result.TopList.Data.Stats.Page*result.TopList.Data.Stats.PageSize < result.TopList.Data.Stats.TotalAssets {
		nextPage, err := page.New(pagination.Number()+1, pagination.RowsPerPage())
		if err != nil {
			return Price{}, err
		}

		return b.AssetPrice(ctx, symbol, nextPage)
	}

	return Price{}, fmt.Errorf("asset %s not found in top list", symbol)
}
