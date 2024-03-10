package config

import (
	"flag"
	"time"
)

const defaultFlushInterval = 300

// ParseFlag parses command line flags and populates the Config struct accordingly.
func ParseFlag(c *Config) {
	var i int
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.StringVar(&c.LogLevel, "l", "info", "Configure the server's log level")
	flag.StringVar(&c.StoreConfig.FileStoragePath, "f", "/tmp/metrics-db.json", "File path to store metrics")
	flag.IntVar(&i, "i", defaultFlushInterval,
		"Time interval in seconds to flush metrics to file, if set to '0' it will flush synchro")
	flag.BoolVar(&c.StoreConfig.Restore, "r", true, "If true metrics will be restored from file path")
	flag.StringVar(&c.StoreConfig.DatabaseDsn, "d", "", "Set postgres DSN")
	flag.StringVar(&c.Key, "k", "", "Set key for hash function")
	flag.Parse()
	c.StoreConfig.StoreInterval = time.Duration(i) * time.Second
}
