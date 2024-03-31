// Package config provides configuration settings for API server.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
	storeConf "github.com/ospiem/mcollector/internal/storage/config"
)

// Config represents the server configuration settings.
type Config struct {
	Endpoint    string           `env:"ADDRESS"`    // Endpoint is the server address.
	CryptoKey   string           `env:"CRYPTO_KEY"` // CryptoKey is used to decrypt the request
	LogLevel    string           `env:"LOG_LEVEL"`  // LogLevel is the logging level.
	Key         string           `env:"KEY"`        // Key is used for hashing func.
	StoreConfig storeConf.Config // StoreConfig holds configuration for storage.
}
type tmpDurations struct {
	StoreInterval int `env:"STORE_INTERVAL"`
}

// New creates new instance of Config by parsing environment variables and command-line flags.
func New() (Config, error) {
	tmp := tmpDurations{StoreInterval: -1}
	var c Config
	ParseFlag(&c)

	err := env.Parse(&tmp)
	if err != nil {
		wrapErr := fmt.Errorf("parse tmp error: %w", err)
		return c, wrapErr
	}

	err = env.Parse(&c)
	if err != nil {
		wrapErr := fmt.Errorf("new server config error: %w", err)
		return c, wrapErr
	}

	if tmp.StoreInterval > 0 {
		c.StoreConfig.StoreInterval = time.Duration(tmp.StoreInterval) * time.Second
	}

	return c, nil
}
