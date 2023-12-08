package transport

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"

	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/rs/zerolog"
)

func TestUpdateTheMetric(t *testing.T) {
	// Create a mock API instance with a mock storage
	storage, err := memorystorage.New()
	if err != nil {
		log.Fatal(err)
	}
	mockAPI := &API{Storage: storage, Log: zerolog.Logger{}, Cfg: config.Config{}}

	tests := []struct {
		name       string
		url        string
		method     string
		body       string
		statusCode int
	}{
		{
			name:       "BadRequestInvalidGaugeValue",
			url:        "/update/gauge/myGauge/invalidValue",
			method:     "POST",
			body:       "",
			statusCode: http.StatusBadRequest,
		},

		// Test Case 4: Bad request (invalid value for counter)
		{
			name:       "BadRequestInvalidCounterValue",
			url:        "/update/counter/myCounter/invalidValue",
			method:     "POST",
			body:       "",
			statusCode: http.StatusBadRequest,
		},

		// Test Case 5: Not found (invalid metric type)
		{
			name:       "NotFoundInvalidMetricType",
			url:        "/update/unknownType/myMetric/42.0",
			method:     "POST",
			body:       "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the specified URL, method, and body
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a mock response recorder
			w := httptest.NewRecorder()

			// Call the handler function
			handler := UpdateTheMetric(mockAPI)
			handler(w, req)

			// Check the response status code
			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
