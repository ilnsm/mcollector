package config

import (
	"flag"
	"time"
)

const defaultFlushInterval = 300

// ParseFlag parses command line flags and populates the Config struct accordingly.
func ParseFlag(c *Config) {
	var i int
	if flag.Lookup("a") == nil {
		flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	}
	if flag.Lookup("l") == nil {
		flag.StringVar(&c.LogLevel, "l", "error", "Configure the server's log level")
	}
	if flag.Lookup("f") == nil {
		flag.StringVar(&c.StoreConfig.FileStoragePath, "f", "/tmp/metrics-db.json", "File path to store metrics")
	}
	if flag.Lookup("i") == nil {
		flag.IntVar(&i, "i", defaultFlushInterval,
			"Time interval in seconds to flush metrics to file, if set to '0' it will flush synchro")
	}
	if flag.Lookup("r") == nil {
		flag.BoolVar(&c.StoreConfig.Restore, "r", true, "If true metrics will be restored from file path")
	}
	if flag.Lookup("d") == nil {
		flag.StringVar(&c.StoreConfig.DatabaseDsn, "d", "", "Set postgres DSN")
	}
	if flag.Lookup("k") == nil {
		flag.StringVar(&c.Key, "k", "", "Set key for hash function")
	}
	if flag.Lookup("crypto-key") == nil {
		flag.StringVar(&c.CryptoKey, "crypto-key", "", "define the private key")
	}

	flag.Parse()
	c.StoreConfig.StoreInterval = time.Duration(i) * time.Second
}
