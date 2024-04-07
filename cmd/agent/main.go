package main

import (
	"os"

	"github.com/ospiem/mcollector/internal/agent"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	if err := agent.Run(logger); err != nil {
		logger.Fatal().Err(err)
	}
	logger.Info().Msg("Graceful shutdown completed successfully. All connections closed, and resources released.")
}
