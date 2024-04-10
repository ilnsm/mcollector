package config

import (
	"flag"
	"time"
)

const defaultReportInterval = 10
const defaultPollInterval = 2

// ParseFlag parses command line flags and populates the Config struct accordingly.
func ParseFlag(c *Config) {
	var ri, pi int
	if flag.Lookup("a") == nil {
		flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	}
	if flag.Lookup("r") == nil {
		flag.IntVar(&ri, "r", defaultReportInterval, "Configure the agent's report interval")
	}
	if flag.Lookup("p") == nil {
		flag.IntVar(&pi, "p", defaultPollInterval, "Configure the agent's poll interval")
	}
	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Set key for hash function")
	}
	if flag.Lookup("log") == nil {
		flag.StringVar(&c.LogLevel, "log", "error", "Configure the agent's log level")
	}
	if flag.Lookup("l") == nil {
		flag.IntVar(&c.RateLimit, "l", 1, "define the number of workers to send metrics")
	}
	if flag.Lookup("crypto-key") == nil {
		flag.StringVar(&c.CryptoKey, "crypto-key", "", "define the public key")
	}
	if flag.Lookup("config") == nil {
		flag.StringVar(&c.Config, "config", "", "define the config file in JSON format")
	}
	flag.Parse()

	c.ReportInterval = time.Duration(ri) * time.Second
	c.PollInterval = time.Duration(pi) * time.Second
}
