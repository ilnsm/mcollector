package main

import (
	"os"

	"github.com/ilnsm/mcollector/internal/server"
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
	api := server.New(cfg, storage)

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
