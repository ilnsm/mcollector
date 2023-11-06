package server

import (
	"fmt"
	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/ilnsm/mcollector/internal/server/transport"
	"github.com/ilnsm/mcollector/internal/storage"
	"github.com/rs/zerolog"
	"net/http"
)

type API struct {
	cfg     config.Config
	storage storage.Storager
	log     zerolog.Logger
}

func New(cfg config.Config, l zerolog.Logger, s storage.Storager) *API {
	return &API{
		cfg:     cfg,
		storage: s,
		log:     l,
	}
}

func (a *API) Run() error {

	fmt.Println("Start server on", a.cfg.Endpoint)
	return http.ListenAndServe(a.cfg.Endpoint, transport.MetrRouter(a.storage))
}
