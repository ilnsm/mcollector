package transport_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/transport"
	"github.com/ospiem/mcollector/internal/storage"
	"github.com/ospiem/mcollector/internal/tools"
	"github.com/rs/zerolog"
)

func ExampleUpdateTheMetricWithJSON() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	tools.SetGlobalLogLevel(cfg.LogLevel)

	ctx := context.Background()

	s, err := storage.New(ctx, cfg.StoreConfig)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	api := transport.New(cfg, s, logger)
	w := httptest.NewRecorder()
	reader := strings.NewReader(`{"id": "gauge_bar",
					"type": "gauge",
					"value": 38.988}`)
	body := io.NopCloser(reader)

	// Generate requset to insert the metric
	request := httptest.NewRequest(http.MethodPost, "/update", body)
	request.Header.Set("Content-Type", "application/json")

	handlerUpdate := transport.UpdateTheMetricWithJSON(api)
	handlerUpdate.ServeHTTP(w, request)
	fmt.Println(w.Body)
	fmt.Println(w.Code)

	// Output: {"value":38.988,"id":"gauge_bar","type":"gauge"}
	//
	// 200
}
