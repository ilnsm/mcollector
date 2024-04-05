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

// JSONConfig represents the configuration settings in JSON format.
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
// It parses the environment variables and the configuration file (if provided).
func New() (Config, error) {
	tmp := tmpDurations{
		ReportInterval: -1,
		PollInterval:   -1,
	}
	var c Config
	ParseFlag(&c)

	// Parse the environment variables into the temporary and main configuration structs
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

	// Convert the temporary durations to time.Duration and assign them to the main configuration
	if tmp.PollInterval > 0 {
		c.ReportInterval = time.Duration(tmp.ReportInterval) * time.Second
	}
	if tmp.ReportInterval > 0 {
		c.PollInterval = time.Duration(tmp.PollInterval) * time.Second
	}

	// Parse the configuration file (if provided)
	err = c.parseConfigFileJSON()
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return c, nil
}

// parseConfigFileJSON parses the configuration file and updates the configuration settings.
// It only updates a setting if it has not been set by an environment variable.
func (c *Config) parseConfigFileJSON() error {
	if c.Config == "" {
		return nil
	}

	// Open the configuration file
	f, err := os.Open(c.Config)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("failed to close the file: %v", closeErr)
		}
	}()

	// Read the configuration file
	confFile, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the configuration file into a temporary JSON configuration struct
	tmp := JSONConfig{}
	err = json.Unmarshal(confFile, &tmp)
	if err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	// Update the main configuration settings with the settings from the configuration file
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
