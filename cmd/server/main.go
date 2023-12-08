package main

import (
	"os"

	"github.com/ilnsm/mcollector/internal/storage"

	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/transport"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	storage, err := storage.New(cfg.FileStoragePath)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	setLogLevel(cfg.LogLevel)
	api := transport.New(cfg, storage, logger)

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
