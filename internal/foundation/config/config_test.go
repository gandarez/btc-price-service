package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gandarez/btc-price-service/internal/foundation/config"
)

func TestConfig_Load(t *testing.T) {
	tmpDir := t.TempDir()

	copyFile(t, "testdata/env", filepath.Join(tmpDir, ".env"))

	cfg, err := config.Load(filepath.Join(tmpDir, ".env"))
	require.NoError(t, err)

	assert.Equal(t, config.Config{
		Environment:     "development",
		ServiceName:     "btc-price-service",
		ShutdownTimeout: 20,
		BroadcastConfig: config.Broadcast{
			MaxPeersPerBroadcaster: 300,
		},
		CacheConfig: config.Cache{
			TTL:                900,
			MaxSize:            50,
			ExpirationInterval: 20,
		},
		CoinDeskConfig: config.CoinDesk{
			URL:          "https://data-api.coindesk.com",
			APIKey:       "some-api-key",
			PollInterval: 10,
		},
		ServerConfig: config.Server{
			Port:              8081,
			ReadHeaderTimeout: 15,
		},
	}, cfg)
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
