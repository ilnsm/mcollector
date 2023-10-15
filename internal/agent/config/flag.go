package config

import (
	"flag"
	"time"
)

func ParseFlag(c *Config) {

	var r, p int
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.IntVar(&r, "r", 10, "Configure the agent's report interval")
	flag.IntVar(&p, "p", 2, "Configure the agent's poll interval")
	flag.Parse()
	c.PollInterval = time.Duration(p) * time.Second
	c.ReportInterval = time.Duration(r) * time.Second

}
