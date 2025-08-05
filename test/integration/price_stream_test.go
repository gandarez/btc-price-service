//go:build integration

package integration_test

import (
	"bufio"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

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
			assert.NotEmpty(t, after, "Data should not be empty")
			assert.JSONEq(t, `{"symbol":"BTC","timestamp":"2025-08-05T17:35:51Z","price":113429.342630482}`, after)

			break // Exit after the first valid data line
		}

		t.Errorf("no valid data received from the stream")
	}
}
