// Package transport provides functionality for handling HTTP transport layer.
package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/ospiem/mcollector/internal/server/middleware/compress"
	"github.com/ospiem/mcollector/internal/server/middleware/hash"
	"github.com/ospiem/mcollector/internal/server/middleware/ssl"
	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/middleware/logger"
	"github.com/rs/zerolog/log"
)

// Storage defines an interface to interact with storage.
//
//go:generate mockgen -destination=../../mock/mock_storage.go  -source=server.go Storage
type Storage interface {
	// InsertGauge inserts a gauge metric into storage.
	InsertGauge(ctx context.Context, k string, v float64) error
	// InsertCounter inserts a counter metric into storage.
	InsertCounter(ctx context.Context, k string, v int64) error
	// SelectGauge selects a gauge metric from storage.
	SelectGauge(ctx context.Context, k string) (float64, error)
	// SelectCounter selects a counter metric from storage.
	SelectCounter(ctx context.Context, k string) (int64, error)
	// GetCounters retrieves all counter metrics from storage.
	GetCounters(ctx context.Context) (map[string]int64, error)
	// GetGauges retrieves all gauge metrics from storage.
	GetGauges(ctx context.Context) (map[string]float64, error)
	// InsertBatch inserts a batch of metrics into storage.
	InsertBatch(ctx context.Context, metrics []models.Metrics) error
	// Ping pings the storage to check its connectivity.
	Ping(ctx context.Context) error
}

// API represents an HTTP API server.
type API struct {
	Storage Storage        // Storage is the storage interface implemention.
	Log     zerolog.Logger // Log is the logger instance.
	Cfg     config.Config  // Cfg is the server configuration.
}

// New creates new instance of API server.
func New(cfg config.Config, s Storage, l zerolog.Logger) *API {
	return &API{
		Cfg:     cfg,
		Storage: s,
		Log:     l,
	}
}

// Run starts the HTTP server.
func (a *API) Run() error {
	log.Info().Msgf("Starting server on %s", a.Cfg.Endpoint)

	r := a.registerAPI()
	if err := http.ListenAndServe(a.Cfg.Endpoint, r); err != nil {
		return fmt.Errorf("run server error: %w", err)
	}
	return nil
}

func (a *API) registerAPI() chi.Router {
	privateKey, err := ssl.ParsePrivateKey(a.Cfg.CryptoKey)
	if err != nil {
		a.Log.Fatal().Msg("failed to parse private key")
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(logger.RequestLogger(a.Log))

	// Mount profiler endpoint for debugging purposes.
	r.Mount("/debug", middleware.Profiler())

	// Define routes for updating metrics.
	r.Route("/update", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// Middleware stack for updating metrics.
			r.Use(compress.DecompressRequest(a.Log))
			r.Use(hash.VerifyRequestBodyIntegrity(a.Log, a.Cfg.Key))
			r.Use(ssl.Terminate(a.Log, privateKey))
			r.Use(compress.CompressResponse(a.Log))

			r.Post("/", UpdateTheMetricWithJSON(a))
			r.Post("/{mType}/{mName}/{mValue}", UpdateTheMetric(a))
			// Route for updating a slice of metrics.
			r.Post("/updates/", UpdateSliceOfMetrics(a))
		})
	})

	r.Group(func(r chi.Router) {
		// Middleware stack for getting metrics.
		r.Use(compress.DecompressRequest(a.Log))

		// Route for listing all metrics.
		r.Get("/", ListAllMetrics(a))

		// Define routes for getting metric values.
		r.Route("/value", func(r chi.Router) {
			r.Post("/", GetTheMetricWithJSON(a))
			r.Get("/{mType}/{mName}", GetTheMetric(a))
		})

		// Route for pinging the database.
		r.Get("/ping", PingDB(a))
	})

	return r
}
