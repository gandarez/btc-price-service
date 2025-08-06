//go:build integration

package integration_test

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/gandarez/btc-price-service/internal/app/domain/priceapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceStream(t *testing.T) {
	apiURL := os.Getenv("BTC_PRICE_STREAM_API_URL")
	url := apiURL + "/v1/price-stream"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read the stream line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if slices.ContainsFunc([]string{": connected", ": ping"}, func(s string) bool {
			return strings.TrimSpace(line) == s
		}) {
			continue
		}

		if after, ok := strings.CutPrefix(line, "data: "); ok {
			require.NotEmpty(t, after, "Data should not be empty")

			var price priceapp.Price

			err := json.Unmarshal([]byte(after), &price)
			require.NoError(t, err)

			assert.Equal(t, price.Symbol, "BTC")
			assert.NotEmpty(t, price.Timestamp)
			assert.NotZero(t, price.Price)

			break // Exit after the first valid data line
		}

		t.Errorf("no valid data received from the stream")
	}
}
