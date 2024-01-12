package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ospiem/mcollector/internal/agent"
	"github.com/ospiem/mcollector/internal/agent/config"
	"github.com/ospiem/mcollector/internal/tools"
	"github.com/rs/zerolog"
)

const timeoutShutdown = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("bye-bye")
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tools.SetLogLevel(cfg.LogLevel)
	logger.Info().Msgf("Start server\nPush to %s\nCollecting metrics every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		logger.Fatal().Msg("failed to gracefully shutdown the service")
	})

	wg := &sync.WaitGroup{}
	defer func() {
		// при выходе из функции ожидаем завершения компонентов приложения
		wg.Wait()
	}()

	wg.Add(1)
	dataChan := agent.Generator(ctx, wg, cfg, logger)

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go agent.Worker(ctx, wg, cfg, dataChan, logger)
	}

	<-ctx.Done()
	return nil
}
