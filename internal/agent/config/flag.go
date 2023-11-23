package config

import (
	"flag"
	"time"
)

func ParseFlag(c *Config) {
	var ri, pi int
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.IntVar(&ri, "r", 10, "Configure the agent's report interval")
	flag.IntVar(&pi, "p", 2, "Configure the agent's poll interval")
	flag.Parse()

	c.ReportInterval = time.Duration(ri) * time.Second
	c.PollInterval = time.Duration(pi) * time.Second
}
