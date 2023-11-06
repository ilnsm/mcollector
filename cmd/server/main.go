package main

import (
	"github.com/ilnsm/mcollector/internal/server"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/storage/memory"
	"github.com/rs/zerolog"
)

func main() {

	logger := zerolog.Logger{}

	storage, err := memorystorage.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	api := server.New(cfg, logger, storage)

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
