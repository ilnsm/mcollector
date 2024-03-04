package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/ospiem/mcollector/internal/server/middleware/compress"
	"github.com/ospiem/mcollector/internal/server/middleware/hash"
	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/middleware/logger"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -destination=../../mock/mock_storage.go  -source=server.go Storage
type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
	GetGauges(ctx context.Context) (map[string]float64, error)
	InsertBatch(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

type API struct {
	Storage Storage
	Log     zerolog.Logger
	Cfg     config.Config
}

func New(cfg config.Config, s Storage, l zerolog.Logger) *API {
	return &API{
		Cfg:     cfg,
		Storage: s,
		Log:     l,
	}
}

func (a *API) Run() error {
	log.Info().Msgf("Starting server on %s", a.Cfg.Endpoint)

	r := a.registerAPI()
	if err := http.ListenAndServe(a.Cfg.Endpoint, r); err != nil {
		return fmt.Errorf("run server error: %w", err)
	}
	return nil
}

func (a *API) registerAPI() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(compress.DecompressRequest(a.Log))
	r.Use(hash.VerifyRequestBodyIntegrity(a.Log, a.Cfg.Key))
	r.Use(logger.RequestLogger(a.Log))
	r.Use(compress.CompressResponse(a.Log))

	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateTheMetricWithJSON(a))
		r.Post("/{mType}/{mName}/{mValue}", UpdateTheMetric(a))
	})

	r.Post("/updates/", UpdateSliceOfMetrics(a))

	r.Get("/", ListAllMetrics(a))

	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetTheMetricWithJSON(a))
		r.Get("/{mType}/{mName}", GetTheMetric(a))
	})

	r.Get("/ping", PingDB(a))

	return r
}
