package coindeskclient

import (
	"net/http"
	"time"
)

const (
	// DefaultTimeoutSecs is the default timeout used for requests to the CoinDesk API.
	DefaultTimeoutSecs = 3
)

// Client communicates with the CoinDesk api.
type Client struct {
	baseURL string
	client  *http.Client
	doFunc  func(c *Client, req *http.Request) (*http.Response, error)
}

// NewClient initializes a new CoinDesk client with the provided base url and api key.
func NewClient(baseURL, apikey string) *Client {
	c := &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: NewTransport(),
		},
		doFunc: func(c *Client, req *http.Request) (*http.Response, error) {
			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Api-Key", apikey)

			return c.client.Do(req)
		},
	}

	return c
}

// Do executes c.doFunc(), which in turn allows wrapping c.client.Do() and manipulating
// the request behavior of the api client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.doFunc(c, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewTransport initializes a new http.Transport.
func NewTransport() *http.Transport {
	return &http.Transport{
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: DefaultTimeoutSecs * time.Second,
	}
}
