package config

import (
	"fmt"
	"github.com/caarlos0/env/v9"
)

type Config struct {
	Endpoint string `env:"ADDRESS"`
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
