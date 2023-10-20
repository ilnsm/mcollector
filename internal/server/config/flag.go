package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v9"
)

func ParseFlag(c *Config) {

	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.Parse()
	err := env.Parse(&c)
	if err != nil {
		fmt.Println(err)
	}
}
