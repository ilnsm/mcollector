package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ospiem/mcollector/internal/helper"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/transport"
	"github.com/ospiem/mcollector/internal/storage"
	"github.com/rs/zerolog"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit)

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	helper.SetGlobalLogLevel(cfg.LogLevel)

	ctx := context.Background()

	s, err := storage.New(ctx, cfg.StoreConfig)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	api := transport.New(cfg, s, logger)

	if err := api.Run(); err != nil {
		logger.Fatal().Err(err).Send()
	}
}
