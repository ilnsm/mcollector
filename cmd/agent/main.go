package main

import (
	"github.com/ospiem/mcollector/internal/agent"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := agent.Run(); err != nil {
		log.Fatal().Err(err).Msg("cannot start agent")
	}
}
