package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/ospiem/mcollector/internal/agent"
	"github.com/ospiem/mcollector/internal/agent/config"
	"github.com/ospiem/mcollector/internal/tools"
	"github.com/rs/zerolog"
)

const timeoutShutdown = 15 * time.Second

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if err := run(logger); err != nil {
		logger.Fatal().Err(err)
	}
	logger.Info().Msg("Graceful shutdown completed successfully. All connections closed, and resources released.")
}

func run(logger zerolog.Logger) error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	tools.SetGlobalLogLevel(cfg.LogLevel)
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

	mc := agent.NewMetricsCollection()
	collectTicker := time.NewTicker(cfg.PollInterval)
	sendTicker := time.NewTicker(cfg.ReportInterval)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	jobs := make(chan map[string]string, cfg.RateLimit)

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go agent.Worker(ctx, wg, cfg, jobs, logger)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-collectTicker.C:
			metrics, err := agent.GetMetrics()
			if err != nil {
				logger.Error().Err(err).Msg("cannot get metrics")
				continue
			}
			mc.Push(metrics)
		case <-sendTicker.C:
			metrics, ok := mc.Pop()
			if !ok {
				continue
			}
			select {
			case jobs <- metrics:
			default:
				logger.Error().Msg("failed to send another job to workers, all workers are busy")
			}
		}
	}
}
