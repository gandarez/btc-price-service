package coindeskclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
)

// TopList calls the CoinDesk API to retrieve the top list of assets.
func (c *Client) TopList(ctx context.Context, page page.Page) (Result, error) {
	url := fmt.Sprintf(
		"%s/asset/v1/top/list?page=%d&page_size=%d&sort_by=CIRCULATING_MKT_CAP_USD&"+
			"sort_direction=DESC&groups=ID,BASIC,PRICE&toplist_quote_asset=BTC",
		c.baseURL,
		page.Number(),
		page.RowsPerPage(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("failed making request to %q: %v", url, err)
	}

	defer resp.Body.Close() // nolint:errcheck,gosec

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("failed reading response body from %q: %v", url, err)
	}

	// 200
	if resp.StatusCode == http.StatusOK {
		topList, err := ParseTopListResponse(body)
		if err != nil {
			return Result{}, err
		}

		return Result{
			TopList: topList,
		}, nil
	}

	// 400 - 503
	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode <= http.StatusServiceUnavailable {
		errMsg, err := ParseTopListResponseError(body)
		if err != nil {
			return Result{}, err
		}

		return Result{
			Error: errMsg,
		}, nil
	}

	return Result{}, fmt.Errorf(
		"invalid response status from %q. got: %d, want: %d. body: %q",
		url,
		resp.StatusCode,
		http.StatusOK,
		string(body),
	)
}

// ParseTopListResponse parses the response from the top/list endpoint.
func ParseTopListResponse(data []byte) (TopList, error) {
	var topList TopList
	if err := json.Unmarshal(data, &topList); err != nil {
		return TopList{}, fmt.Errorf("failed to parse top list response: %v", err)
	}

	return topList, nil
}

// ParseTopListResponseError parses the error response from the top/list endpoint.
func ParseTopListResponseError(data []byte) (string, error) {
	type responseBodyErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"Err"`
	}

	var errResp responseBodyErr
	if err := json.Unmarshal(data, &errResp); err != nil {
		return "", fmt.Errorf("failed to parse error response: %v", err)
	}

	return errResp.Error.Message, nil
}
