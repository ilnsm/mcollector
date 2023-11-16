package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/middleware/logger"
	"github.com/ilnsm/mcollector/internal/server/transport"
	"github.com/ilnsm/mcollector/internal/storage"
	"github.com/rs/zerolog/log"
)

type API struct {
	cfg     config.Config
	storage storage.Storager
}

func New(cfg config.Config, s storage.Storager) *API {
	return &API{
		cfg:     cfg,
		storage: s,
	}
}

func (a *API) Run() error {
	log.Info().Msgf("Starting server on %s", a.cfg.Endpoint)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(transport.CheckMetricType)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{gName}/{gValue}", transport.UpdateGauge(a.storage))
		r.Post("/counter/{cName}/{cValue}", transport.UpdateCounter(a.storage))
	})

	r.Get("/", transport.ListAllMetrics(a.storage))

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{gName}", transport.GetGauge(a.storage))
		r.Get("/counter/{cName}", transport.GetCounter(a.storage))
	})

	return http.ListenAndServe(a.cfg.Endpoint, r)
}
