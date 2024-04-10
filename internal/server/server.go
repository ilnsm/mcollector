package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ospiem/mcollector/internal/helper"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/transport"
	"github.com/ospiem/mcollector/internal/storage"
	"github.com/rs/zerolog"
)

// buildVersion, buildDate, and buildCommit are variables that hold the build information.
// They are set at build time using ldflags.
var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

// timeoutShutdown is the duration to wait for the server to shut down gracefully.
const timeoutShutdown = 5 * time.Second

// Run is the main function of the server. It initializes the server and its components,
// and manages their lifecycle.
func Run(logger zerolog.Logger) error {
	// Create a context that is cancelled when an interrupt signal is received.
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelCtx()

	// If the service fails to shut down gracefully, log a fatal error.
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		logger.Fatal().Msg("failed to gracefully shutdown the service")
	})

	// Log the build information.
	logger.Log().
		Str("Build version", buildVersion).
		Str("Build date", buildDate).
		Str("Build commit", buildCommit).
		Msg("Starting server")

	// Initialize the server configuration.
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Set the global log level.
	helper.SetGlobalLogLevel(cfg.LogLevel)

	// Initialize a WaitGroup to wait for the completion of application components.
	wg := &sync.WaitGroup{}
	defer func() {
		// When exiting the main function, we expect the completion of application components
		wg.Wait()
	}()

	// Initialize the s.
	s, err := storage.New(ctx, cfg.StoreConfig)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to initialize s: %w", err)
	}

	// Watch the s for closure.
	watchStorage(ctx, wg, s, &logger)
	// Initialize the API and the server.
	componentsErrs := make(chan error, 1)
	api := transport.New(&cfg, s, &logger)
	srv := api.InitServer()

	// Manage the server lifecycle.
	manageServer(ctx, wg, srv, componentsErrs, &logger)

	// Wait for the context to be done or for an error to occur.
	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		logger.Error().Err(err)
		cancelCtx()
	}

	return nil
}

// watchStorage watches the storage for closure and logs any errors that occur during closure.
func watchStorage(ctx context.Context, wg *sync.WaitGroup, s transport.Storage, l *zerolog.Logger) {
	wg.Add(1)
	go func() {
		defer l.Info().Msg("Storage has been closed")
		defer wg.Done()

		<-ctx.Done()

		if err := s.Close(ctx); err != nil {
			l.Error().Err(err).Msg("failed to close storage")
		}
	}()
}

// manageServer manages the lifecycle of the server. It starts the server and handles shutdown.
func manageServer(ctx context.Context, wg *sync.WaitGroup, srv *http.Server, errs chan error, l *zerolog.Logger) {
	// Start the server in a separate goroutine.
	go func(errs chan<- error) {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and serve has failed: %w", err)
		}
	}(errs)

	// Handle server shutdown in a separate goroutine.
	wg.Add(1)
	go func() {
		defer l.Info().Msg("Server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutDownTimeoutCtx, cancelShutdownTimeCancel := context.WithTimeout(ctx, timeoutShutdown)
		defer cancelShutdownTimeCancel()
		if err := srv.Shutdown(shutDownTimeoutCtx); err != nil {
			l.Error().Err(err).Msg("an error occurred during server shutdown")
		}
	}()
}
