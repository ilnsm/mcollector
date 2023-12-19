package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	Endpoint        string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   time.Duration
	Database_DSN    string `env:"DATABASE_DSN"`
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
		c.StoreInterval = time.Duration(tmp.StoreInterval) * time.Second
	}

	return c, nil
}
