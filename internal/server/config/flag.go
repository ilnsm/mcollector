package config

import (
	"flag"
	"time"
)

const defaultFlushInterfal = 300

func ParseFlag(c *Config) {
	var i int
	flag.StringVar(&c.Endpoint, "a", "localhost:8080", "Configure the server's host:port")
	flag.StringVar(&c.LogLevel, "l", "info", "Configure the server's log level")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "File path to store metrics")
	flag.IntVar(&i, "i", defaultFlushInterfal,
		"Time interval in seconds to flush metrics to file, if set to '0' it will flush synchro")
	flag.BoolVar(&c.Restore, "r", true, "If true metrics will be restored from file path")
	flag.StringVar(&c.DatabaseDsn, "d", "", "postgres dns")
	flag.Parse()
	c.StoreInterval = time.Duration(i) * time.Second
}
