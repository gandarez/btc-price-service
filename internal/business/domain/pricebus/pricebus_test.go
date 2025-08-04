package pricebus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gandarez/btc-price-service/internal/business/domain/pricebus"
	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
)

func TestBusiness_AssetPrice(t *testing.T) {
	mockCoinDeskClient := &mockCoinDeskClient{
		TopListFn: func(_ context.Context, _ page.Page) (coindeskclient.Result, error) {
			return coindeskclient.Result{
				TopList: coindeskclient.TopList{
					Data: coindeskclient.Data{
						Stats: coindeskclient.Stats{
							Page:        1,
							PageSize:    10,
							TotalAssets: 1,
						},
						Assets: []coindeskclient.Asset{
							{Symbol: "BTC", Price: 50000.0, PriceLastUpdatedAt: 1633072800},
						},
					},
				},
			}, nil
		},
	}

	bus := pricebus.NewBusiness(mockCoinDeskClient)

	page, err := page.New(1, 10)
	require.NoError(t, err)

	result, err := bus.AssetPrice(t.Context(), "BTC", page)
	require.NoError(t, err)

	assert.Equal(t, pricebus.Price{
		Symbol:    "BTC",
		Timestamp: "2021-10-01T07:20:00Z",
		Price:     50000.0,
	}, result)
	assert.Equal(t, 1, mockCoinDeskClient.TopListFnCount)
}

func TestBusiness_AssetPrice_Recursive(t *testing.T) {
	mockCoinDeskClient := &mockCoinDeskClient{
		TopListFn: func(_ context.Context, p page.Page) (coindeskclient.Result, error) {
			if p.Number() == 1 {
				return coindeskclient.Result{
					TopList: coindeskclient.TopList{
						Data: coindeskclient.Data{
							Stats: coindeskclient.Stats{
								Page:        1,
								PageSize:    1,
								TotalAssets: 2,
							},
							Assets: []coindeskclient.Asset{
								{Symbol: "ETH", Price: 3000.0, PriceLastUpdatedAt: 1633072800},
							},
						},
					},
				}, nil
			}

			return coindeskclient.Result{
				TopList: coindeskclient.TopList{
					Data: coindeskclient.Data{
						Stats: coindeskclient.Stats{
							Page:        2,
							PageSize:    1,
							TotalAssets: 1,
						},
						Assets: []coindeskclient.Asset{
							{Symbol: "BTC", Price: 50000.0, PriceLastUpdatedAt: 1633072800},
						},
					},
				},
			}, nil
		},
	}

	bus := pricebus.NewBusiness(mockCoinDeskClient)

	page, err := page.New(1, 1)
	require.NoError(t, err)

	result, err := bus.AssetPrice(t.Context(), "BTC", page)
	require.NoError(t, err)

	assert.Equal(t, pricebus.Price{
		Symbol:    "BTC",
		Timestamp: "2021-10-01T07:20:00Z",
		Price:     50000.0,
	}, result)
	assert.Equal(t, 2, mockCoinDeskClient.TopListFnCount)
}

func TestBusiness_AssetPrice_Err(t *testing.T) {
	mockCoinDeskClient := &mockCoinDeskClient{
		TopListFn: func(_ context.Context, _ page.Page) (coindeskclient.Result, error) {
			return coindeskclient.Result{}, errors.New("fail")
		},
	}

	bus := pricebus.NewBusiness(mockCoinDeskClient)

	page, err := page.New(1, 10)
	require.NoError(t, err)

	_, err = bus.AssetPrice(t.Context(), "BTC", page)

	assert.EqualError(t, err, "failed to fetch top list: fail")
	assert.Equal(t, 1, mockCoinDeskClient.TopListFnCount)
}

func TestBusiness_AssetPrice_APIErr(t *testing.T) {
	mockCoinDeskClient := &mockCoinDeskClient{
		TopListFn: func(_ context.Context, _ page.Page) (coindeskclient.Result, error) {
			return coindeskclient.Result{
				Error: "bad request",
			}, nil
		},
	}

	bus := pricebus.NewBusiness(mockCoinDeskClient)

	page, err := page.New(1, 10)
	require.NoError(t, err)

	_, err = bus.AssetPrice(t.Context(), "BTC", page)

	assert.EqualError(t, err, "failed to fetch top list: bad request")
	assert.Equal(t, 1, mockCoinDeskClient.TopListFnCount)
}

func TestBusiness_AssetPrice_NoAssets(t *testing.T) {
	mockCoinDeskClient := &mockCoinDeskClient{
		TopListFn: func(_ context.Context, _ page.Page) (coindeskclient.Result, error) {
			return coindeskclient.Result{}, nil
		},
	}

	bus := pricebus.NewBusiness(mockCoinDeskClient)

	page, err := page.New(1, 10)
	require.NoError(t, err)

	_, err = bus.AssetPrice(t.Context(), "BTC", page)

	assert.EqualError(t, err, "no assets found for page 1")
	assert.Equal(t, 1, mockCoinDeskClient.TopListFnCount)
}

type mockCoinDeskClient struct {
	TopListFn      func(ctx context.Context, page page.Page) (coindeskclient.Result, error)
	TopListFnCount int
}

func (m *mockCoinDeskClient) TopList(ctx context.Context, p page.Page) (coindeskclient.Result, error) {
	m.TopListFnCount++
	return m.TopListFn(ctx, p)
}
