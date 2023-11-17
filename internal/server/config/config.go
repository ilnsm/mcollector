package config

import (
	"github.com/rs/zerolog/log"

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
		log.Err(err)
		return c, err
	}

	return c, nil
}
