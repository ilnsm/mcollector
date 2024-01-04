package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
	storeConf "github.com/ospiem/mcollector/internal/storage/config"
)

type Config struct {
	Endpoint    string `env:"ADDRESS"`
	LogLevel    string `env:"LOG_LEVEL"`
	Key         string `env:"KEY"`
	StoreConfig storeConf.Config
}
type tmpDurations struct {
	StoreInterval int `env:"STORE_INTERVAL"`
}

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
