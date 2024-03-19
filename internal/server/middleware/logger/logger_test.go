package logger_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/ospiem/mcollector/internal/server/middleware/logger"
)

func TestRequestLogger(t *testing.T) {
	var (
		recordedLog *zerolog.Event
	)

	log := zerolog.Nop().Level(zerolog.DebugLevel).
		Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			recordedLog = e
		}))

	middleware := logger.RequestLogger(log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			log.Error().Err(err).Msg("failed to write response")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.NotNil(t, recordedLog)
}
