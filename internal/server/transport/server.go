package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilnsm/mcollector/internal/server/middleware/compress"
	"github.com/jackc/pgx/v5"

	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/middleware/logger"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) map[string]int64
	GetGauges(ctx context.Context) map[string]float64
	Ping(ctx context.Context) error
}

type API struct {
	Storage Storage
	Conn    *pgx.Conn
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

func (a *API) Run(ctx context.Context) error {
	log.Info().Msgf("Starting server on %s", a.Cfg.Endpoint)

	r := a.registerAPI(ctx)
	if err := http.ListenAndServe(a.Cfg.Endpoint, r); err != nil {
		return fmt.Errorf("run server error: %w", err)
	}
	return nil
}

func (a *API) registerAPI(ctx context.Context) chi.Router {
	r := chi.NewRouter()
	r.Use(compress.DecompressRequest(a.Log))
	r.Use(logger.RequestLogger(a.Log))
	r.Use(compress.CompressResponse(a.Log))

	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateTheMetricWithJSON(ctx, a))
		r.Post("/{mType}/{mName}/{mValue}", UpdateTheMetric(ctx, a))
	})

	r.Get("/", ListAllMetrics(ctx, a))

	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetTheMetricWithJSON(ctx, a))
		r.Get("/{mType}/{mName}", GetTheMetric(ctx, a))
	})

	r.Get("/ping", PingDB(ctx, a))

	return r
}
