package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	Endpoint       string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	LogLevel       string `env:"LOG_LEVEL"`
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type tmpDurations struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

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
