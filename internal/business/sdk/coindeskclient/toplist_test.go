package coindeskclient_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gandarez/btc-price-service/internal/business/sdk/coindeskclient"
	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
)

const apikey = "318xqznbvafjtlcewphudyrkgomsivwxjkqltbedcyrumhzsganowpkjdu7352"

func TestClient_TopList(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/asset/v1/top/list", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{apikey}, req.Header["X-Api-Key"])

		err := req.ParseForm()
		require.NoError(t, err)

		// check query params
		assert.True(t, req.Form.Has("page"))
		assert.True(t, req.Form.Has("page_size"))
		assert.True(t, req.Form.Has("sort_by"))
		assert.True(t, req.Form.Has("sort_direction"))
		assert.True(t, req.Form.Has("groups"))
		assert.True(t, req.Form.Has("toplist_quote_asset"))

		// write response
		f, err := os.Open("testdata/api_toplist_response.json")
		require.NoError(t, err)

		defer f.Close() // nolint:errcheck,gosec

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	page, err := page.New(1, 10)
	require.NoError(t, err)

	c := coindeskclient.NewClient(url, apikey)
	result, err := c.TopList(t.Context(), page)
	require.NoError(t, err)

	assert.Len(t, result.TopList.Data.Assets, 10)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_TopList_ErrBadRequest(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/asset/v1/top/list", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		// write response
		f, err := os.Open("testdata/api_toplist_bad_request_response.json")
		require.NoError(t, err)

		defer f.Close() // nolint:errcheck,gosec

		w.WriteHeader(http.StatusBadRequest)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	page, err := page.New(1, 10)
	require.NoError(t, err)

	c := coindeskclient.NewClient(url, apikey)
	result, err := c.TopList(t.Context(), page)
	require.NoError(t, err)

	assert.Len(t, result.TopList.Data.Assets, 0)
	assert.Equal(t, result.Error, "Not found: market parameter.")

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_TopList_DefaultErr(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/asset/v1/top/list", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusGatewayTimeout)
	})

	page, err := page.New(1, 10)
	require.NoError(t, err)

	c := coindeskclient.NewClient(url, apikey)
	_, err = c.TopList(t.Context(), page)

	assert.Contains(t, err.Error(), "invalid response status from")

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_TopList_InvalidURL(t *testing.T) {
	page, err := page.New(1, 10)
	require.NoError(t, err)

	c := coindeskclient.NewClient("invalid-url", apikey)
	_, err = c.TopList(t.Context(), page)

	assert.Contains(t, err.Error(), "failed making request to")
}

func TestParseTopListResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/api_toplist_simplified_response.json")
	require.NoError(t, err)

	result, err := coindeskclient.ParseTopListResponse(data)
	require.NoError(t, err)

	assert.Equal(t, coindeskclient.TopList{
		Data: coindeskclient.Data{
			Stats: coindeskclient.Stats{
				Page:        1,
				PageSize:    10,
				TotalAssets: 3145,
			},
			Assets: []coindeskclient.Asset{
				{
					ID:                 1,
					Symbol:             "BTC",
					Price:              113907.168087996,
					PriceLastUpdatedAt: 1754218141,
				},
				{
					ID:                 2,
					Symbol:             "ETH",
					Price:              3483.56905401533,
					PriceLastUpdatedAt: 1754218141,
				},
				{
					ID:                 13,
					Symbol:             "XRP",
					Price:              2.87429329041395,
					PriceLastUpdatedAt: 1754218141,
				},
				{
					ID:                 7,
					Symbol:             "USDT",
					Price:              1.00018862042186,
					PriceLastUpdatedAt: 1754218135,
				},
				{
					ID:                 8,
					Symbol:             "BNB",
					Price:              750.607092526658,
					PriceLastUpdatedAt: 1754218141,
				},
				{
					ID:                 3,
					Symbol:             "SOL",
					Price:              161.794159381107,
					PriceLastUpdatedAt: 1754218137,
				},
				{
					ID:                 14,
					Symbol:             "USDC",
					Price:              1.00009464675868,
					PriceLastUpdatedAt: 1754218140,
				},
				{
					ID:                 28,
					Symbol:             "TRX",
					Price:              0.326770605698632,
					PriceLastUpdatedAt: 1754218140,
				},
				{
					ID:                 26,
					Symbol:             "DOGE",
					Price:              0.197819930141338,
					PriceLastUpdatedAt: 1754218140,
				},
				{
					ID:                 12,
					Symbol:             "ADA",
					Price:              0.726014958952661,
					PriceLastUpdatedAt: 1754218140,
				},
			},
		},
	}, result)
}
func TestParseTopListResponseError(t *testing.T) {
	data, err := os.ReadFile("testdata/api_toplist_bad_request_response.json")
	require.NoError(t, err)

	result, err := coindeskclient.ParseTopListResponseError(data)
	require.NoError(t, err)

	assert.Equal(t, "Not found: market parameter.", result)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
