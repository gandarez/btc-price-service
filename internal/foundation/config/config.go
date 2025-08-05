package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type (
	// Config holds the application configuration.
	Config struct {
		Environment     string    `mapstructure:"ENVIRONMENT"`
		ServiceName     string    `mapstructure:"SERVICE_NAME"`
		ShutdownTimeout int       `mapstructure:"SHUTDOWN_TIMEOUT"` // in seconds
		BroadcastConfig Broadcast `mapstructure:",squash"`
		CacheConfig     Cache     `mapstructure:",squash"`
		CoinDeskConfig  CoinDesk  `mapstructure:",squash"`
		ServerConfig    Server    `mapstructure:",squash"`
	}

	// Broadcast holds the configuration for the pubsub broadcaster.
	Broadcast struct {
		MaxPeersPerBroadcaster int `mapstructure:"BROADCAST_MAX_PEERS_PER_BROADCASTER"`
	}

	// Cache holds the configuration for the in-memory cache.
	Cache struct {
		TTL                int `mapstructure:"CACHE_TTL"`                 // time to live for cache entries in seconds
		MaxSize            int `mapstructure:"CACHE_MAX_SIZE"`            // maximum number of entries in the cache
		ExpirationInterval int `mapstructure:"CACHE_EXPIRATION_INTERVAL"` // interval to check for expired entries in seconds
	}

	// CoinDesk holds the configuration for the CoinDesk API.
	CoinDesk struct {
		URL          string `mapstructure:"COINDESK_URL"`
		APIKey       string `mapstructure:"COINDESK_API_KEY"`
		PollInterval int    `mapstructure:"COINDESK_POLL_INTERVAL"` // in seconds
	}

	// Server holds the configuration for the HTTP server.
	Server struct {
		Port              int `mapstructure:"SERVER_PORT"`
		ReadHeaderTimeout int `mapstructure:"SERVER_READ_HEADER_TIMEOUT"`
	}
)

// Load loads the application configuration.
func Load(path string) (Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

// String implements fmt.Stringer interface.
func (b Broadcast) String() string {
	return fmt.Sprintf("max peers per broadcaster: %d", b.MaxPeersPerBroadcaster)
}

// String implements fmt.Stringer interface.
func (c Cache) String() string {
	return fmt.Sprintf("ttl: %d, max size: %d, expiration interval: %d", c.TTL, c.MaxSize, c.ExpirationInterval)
}

// String implements fmt.Stringer interface.
func (cd CoinDesk) String() string {
	return fmt.Sprintf("url: %s, apiKey: %s, poll interval: %d", cd.URL, cd.APIKey, cd.PollInterval)
}

// String implements fmt.Stringer interface.
func (s Server) String() string {
	return fmt.Sprintf("port: %d, read header timeout: %d", s.Port, s.ReadHeaderTimeout)
}

// String implements fmt.Stringer interface.
func (c Config) String() string {
	return fmt.Sprintf("env: %s, service: %s, shutdown timeout: %d,"+
		" broadcast: (%s), cache: (%s), coindesk: (%s), server: (%s)",
		c.Environment, c.ServiceName, c.ShutdownTimeout,
		c.BroadcastConfig, c.CacheConfig, c.CoinDeskConfig, c.ServerConfig,
	)
}
