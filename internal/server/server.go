package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/middleware/logger"
	"github.com/ilnsm/mcollector/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

type API struct {
	cfg     config.Config
	storage storage.Storager
}

func New(cfg config.Config, l zerolog.Logger, s storage.Storager) *API {
	return &API{
		cfg:     cfg,
		storage: s,
	}
}

func (a *API) Run() error {

	log.Info().Msgf("Starting server on %s", a.cfg.Endpoint)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(checkMetricType)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{gName}/{gValue}", updateGauge(a.storage))
		r.Post("/counter/{cName}/{cValue}", updateCounter(a.storage))
	})

	r.Get("/", listAllMetrics(a.storage))

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{gName}", getGauge(a.storage))
		r.Get("/counter/{cName}", getCounter(a.storage))

	})

	return http.ListenAndServe(a.cfg.Endpoint, r)
}
