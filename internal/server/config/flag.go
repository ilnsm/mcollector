package config

import (
	"flag"
)

func ParseFlag(c *Config) {
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.StringVar(&c.LogLevel, "l", "info", "Configure the server's log level")
	flag.Parse()
}
