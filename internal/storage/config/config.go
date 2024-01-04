package config

import "time"

type Config struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   time.Duration
}
