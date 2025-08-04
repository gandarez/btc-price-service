package coindeskclient

type (
	// Result represents the response structure from the CoinDesk API for the top list of assets.
	Result struct {
		Error   string
		TopList TopList
	}

	// TopList represents the top list of assets returned by the CoinDesk API.
	TopList struct {
		Data Data `json:"Data"`
	}

	// Data contains the list of assets in the top list.
	Data struct {
		Stats  Stats   `json:"STATS"`
		Assets []Asset `json:"LIST"`
	}

	// Stats contains pagination and total asset information.
	Stats struct {
		Page        int `json:"PAGE"`
		PageSize    int `json:"PAGE_SIZE"`
		TotalAssets int `json:"TOTAL_ASSETS"`
	}

	// Asset represents an individual asset in the top list.
	Asset struct {
		ID                 int     `json:"ID"`
		Price              float64 `json:"PRICE_USD"`
		PriceLastUpdatedAt int64   `json:"PRICE_USD_LAST_UPDATE_TS"`
		Symbol             string  `json:"SYMBOL"`
	}
)
