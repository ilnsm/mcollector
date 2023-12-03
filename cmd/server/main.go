package main

import (
	"github.com/ilnsm/mcollector/internal/server/transport"
	"github.com/ilnsm/mcollector/internal/storage/file"
	"os"
	"time"

	"github.com/ilnsm/mcollector/internal/server/config"
	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	storage, err := memorystorage.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	setLogLevel(cfg.LogLevel)
	api := transport.New(cfg, storage, logger)

	if cfg.Restore {
		logger.Debug().Msg("append to restore metrics")

		err := file.RestoreMetrics(storage, cfg.FileStoragePath, logger)
		if err != nil {
			logger.Error().Err(err).Msg("cannot restore the data")
		}

		logger.Debug().Msg("restored metrics")
	}

	if cfg.StoreInterval > 0 {
		t := time.NewTicker(cfg.StoreInterval)
		defer t.Stop()

		go func() {
			for {
				select {
				case <-t.C:
					logger.Debug().Msg("attempt to flush metrics by ticker")
					err := file.FlushMetrics(storage, cfg.FileStoragePath)
					if err != nil {
						logger.Error().Err(err).Msg("cannot flush metrics in time")
					}
				}
			}
		}()
	}

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
