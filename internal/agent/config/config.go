package config

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	"time"
)

type Config struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

func New() (Config, error) {
	var c Config
	ParseFlag(&c)

	err := env.Parse(&c)
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	return c, nil
}
