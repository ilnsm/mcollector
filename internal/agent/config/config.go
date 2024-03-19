// Package config provides functionality to manage configuration settings.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

// Config represents the configuration settings.
type Config struct {
	Endpoint       string        `env:"ADDRESS"`   // Endpoint for sending metrics.
	Key            string        `env:"KEY"`       // Key is used for hashing func.
	LogLevel       string        `env:"LOG_LEVEL"` // LogLevel is the logging level.
	ReportInterval time.Duration // Time interval for reporting metrics
	PollInterval   time.Duration // Time interval for polling metrics
	RateLimit      int           `env:"RATE_LIMIT"` // Rate limit for sending metrics
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
	return c, nil
}
