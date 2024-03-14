package transport_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/transport"
	"github.com/ospiem/mcollector/internal/storage"
	"github.com/rs/zerolog"
)

func Example() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	ctx := context.Background()

	s, err := storage.New(ctx, cfg.StoreConfig)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	api := transport.New(cfg, s, logger)

	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		api.Log.Fatal().Err(err).Send()
	}

	rr := httptest.NewRecorder()
	handler := transport.PingDB(api)

	handler.ServeHTTP(rr, req)

	fmt.Println(rr.Code)

	// Output:
	// 200
}
