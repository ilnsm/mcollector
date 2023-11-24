package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	Endpoint       string `env:"ADDRESS"`
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type tmpDurations struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

func New() (Config, error) {
	var tmp tmpDurations
	var c Config
	ParseFlag(&c)

	err := env.Parse(&tmp)
	if err != nil {
		wrapErr := fmt.Errorf("new agent config error: %w", err)
		return c, wrapErr
	}
	err = env.Parse(&c)
	if err != nil {
		wrapErr := fmt.Errorf("new agent config error: %w", err)
		return c, wrapErr
	}

	c.ReportInterval = time.Duration(tmp.ReportInterval) * time.Second
	c.PollInterval = time.Duration(tmp.PollInterval) * time.Second
	return c, nil
}
