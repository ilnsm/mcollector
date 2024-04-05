// Package config provides functionality to manage configuration settings.
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/rs/zerolog/log"
)

// Config represents the configuration settings.
type Config struct {
	Endpoint       string        `env:"ADDRESS"`    // Endpoint for sending metrics.
	CryptoKey      string        `env:"CRYPTO_KEY"` // CRYPTO_KEY is used to encrypt the request
	Config         string        `env:"CONFIG"`     // Config is path to the config file.
	Key            string        `env:"KEY"`        // Key is used for hashing func.
	LogLevel       string        `env:"LOG_LEVEL"`  // LogLevel is the logging level.
	ReportInterval time.Duration // Time interval for reporting metrics
	PollInterval   time.Duration // Time interval for polling metrics
	RateLimit      int           `env:"RATE_LIMIT"` // Rate limit for sending metrics
}

type JSONConfig struct {
	Endpoint       string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// tmpDurations represents temporary durations for parsing environment variables.
type tmpDurations struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

// New creates a new configuration instance.
func New() (Config, error) {
	tmp := tmpDurations{
		ReportInterval: -1,
		PollInterval:   -1,
	}
	var c Config
	ParseFlag(&c)

	err := env.Parse(&tmp)
	if err != nil {
		wrapErr := fmt.Errorf("parse tmp error: %w", err)
		return c, wrapErr
	}

	err = env.Parse(&c)
	if err != nil {
		wrapErr := fmt.Errorf("parse config error: %w", err)
		return c, wrapErr
	}

	if tmp.PollInterval > 0 {
		c.ReportInterval = time.Duration(tmp.ReportInterval) * time.Second
	}
	if tmp.ReportInterval > 0 {
		c.PollInterval = time.Duration(tmp.PollInterval) * time.Second
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
	if c.ReportInterval == defaultReportInterval*time.Second {
		interval, err := time.ParseDuration(tmp.ReportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse store interval: %w", err)
		}
		c.ReportInterval = interval
	}
	if c.PollInterval == defaultPollInterval*time.Second {
		interval, err := time.ParseDuration(tmp.PollInterval)
		if err != nil {
			return fmt.Errorf("failed to parse store interval: %w", err)
		}
		c.PollInterval = interval
	}

	return nil
}
