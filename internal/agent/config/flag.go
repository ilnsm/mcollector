package config

import (
	"flag"
)

func ParseFlag(c *Config) {
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.IntVar(&c.ReportInterval, "r", 10, "Configure the agent's report interval")
	flag.IntVar(&c.PollInterval, "p", 2, "Configure the agent's poll interval")
	flag.Parse()
}
