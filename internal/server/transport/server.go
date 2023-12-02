package transport

import (
	"net/http"

	"github.com/ilnsm/mcollector/internal/server/middleware/compress"

	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/middleware/logger"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	SelectGauge(k string) (float64, error)
	SelectCounter(k string) (int64, error)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
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

	return http.ListenAndServe(a.Cfg.Endpoint, r)
}

func (a *API) registerAPI() chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger(a.Log))
	r.Use(compress.CompressResponse(a.Log))
	r.Use(compress.DecompressRequest(a.Log))

	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateTheMetricWithJSON(a))
		r.Post("/{mType}/{mName}/{mValue}", UpdateTheMetric(a))
	})

	r.Get("/", ListAllMetrics(a))

	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetTheMetricWithJSON(a))
		r.Get("/{mType}/{mName}", GetTheMetric(a))
	})

	return r
}
