package config

import (
	"flag"
	"time"
)

const defaultReportInterval = 10
const defaultPollInterval = 2

func ParseFlag(c *Config) {
	var ri, pi int
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.IntVar(&ri, "r", defaultReportInterval, "Configure the agent's report interval")
	flag.IntVar(&pi, "p", defaultPollInterval, "Configure the agent's poll interval")
	flag.Parse()

	c.ReportInterval = time.Duration(ri) * time.Second
	c.PollInterval = time.Duration(pi) * time.Second
}
