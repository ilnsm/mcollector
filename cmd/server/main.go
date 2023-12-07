package main

import (
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/transport"
	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"
	"github.com/rs/zerolog"
	"os"
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

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
