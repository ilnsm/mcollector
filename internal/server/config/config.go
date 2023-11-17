package config

import (
	"fmt"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	Endpoint string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func New() (Config, error) {
	var c Config
	ParseFlag(&c)

	err := env.Parse(&c)
	if err != nil {
		wrapErr := fmt.Errorf("new server config error: %w", err)
		return c, wrapErr
	}

	return c, nil
}
