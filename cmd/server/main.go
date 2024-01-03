package main

import (
	"context"
	"os"

	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/transport"
	"github.com/ospiem/mcollector/internal/storage"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	setLogLevel(cfg.LogLevel)

	ctx := context.Background()

	s, err := storage.New(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	api := transport.New(cfg, s, logger)

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
