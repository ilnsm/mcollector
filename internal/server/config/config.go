// Package config provides configuration settings for API server.
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v9"
	storeConf "github.com/ospiem/mcollector/internal/storage/config"
	"github.com/rs/zerolog/log"
)

// Config represents the server configuration settings.
type Config struct {
	Endpoint    string           `env:"ADDRESS"`    // Endpoint is the server address.
	Config      string           `env:"CONFIG"`     // Config is path to the config file.
	CryptoKey   string           `env:"CRYPTO_KEY"` // CryptoKey is used to decrypt the request.
	LogLevel    string           `env:"LOG_LEVEL"`  // LogLevel is the logging level.
	Key         string           `env:"KEY"`        // Key is used for hashing func.
	StoreConfig storeConf.Config // StoreConfig holds configuration for storage.
}

type JSONConfig struct {
	Endpoint      string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDsn   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
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

	err = c.parseConfigFileJSON()
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return c, nil
}

func (c *Config) parseConfigFileJSON() error {
	if c.Config == "" {
		return nil
	}

	f, err := os.Open(c.Config)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("failed to close the file: %v", closeErr)
		}
	}()

	confFile, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	tmp := JSONConfig{}
	err = json.Unmarshal(confFile, &tmp)
	if err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	if c.Endpoint == "" {
		c.Endpoint = tmp.Endpoint
	}
	if c.CryptoKey == "" {
		c.CryptoKey = tmp.CryptoKey
	}
	if c.StoreConfig.FileStoragePath == "" {
		c.StoreConfig.FileStoragePath = tmp.StoreFile
	}
	if c.StoreConfig.Restore == false {
		c.StoreConfig.Restore = tmp.Restore
	}
	if c.StoreConfig.StoreInterval == defaultFlushInterval*time.Second {
		interval, err := time.ParseDuration(tmp.StoreInterval)
		if err != nil {
			return fmt.Errorf("failed to parse store interval: %w", err)
		}
		c.StoreConfig.StoreInterval = interval
	}
	if c.StoreConfig.DatabaseDsn == "" {
		c.StoreConfig.DatabaseDsn = tmp.DatabaseDsn
	}

	return nil
}
