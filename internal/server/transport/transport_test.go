package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilnsm/mcollector/internal/server/config"
	"github.com/rs/zerolog"
)

type MockStorage struct {
	// Implement necessary fields for mock storage
}

func (s *MockStorage) InsertGauge(name string, value float64) error {
	return nil
}

func (s *MockStorage) InsertCounter(name string, value int64) error {
	return nil
}
func (s *MockStorage) SelectGauge(k string) (float64, error) {
	return 0, nil
}
func (s *MockStorage) SelectCounter(k string) (int64, error) {
	return 0, nil
}

func (s *MockStorage) GetCounters() map[string]int64 {
	return nil
}
func (s *MockStorage) GetGauges() map[string]float64 {
	return nil
}

func TestUpdateTheMetric(t *testing.T) {
	// Create a mock API instance with a mock storage
	mockStorage := &MockStorage{}
	mockAPI := &API{Storage: mockStorage, Log: zerolog.Logger{}, Cfg: config.Config{}}

	tests := []struct {
		name       string
		url        string
		method     string
		body       string
		statusCode int
	}{
		// Test Case 1: Successful update of a gauge
		//{
		//	name:       "UpdateGaugeSuccess",
		//	url:        "/update/gauge/myGauge/42.0",
		//	method:     "POST",
		//	body:       "",
		//	statusCode: http.StatusOK,
		// },
		//
		//// Test Case 2: Successful update of a counter
		//{
		//	name:       "UpdateCounterSuccess",
		//	url:        "/update/counter/myCounter/10",
		//	method:     "POST",
		//	body:       "",
		//	statusCode: http.StatusOK,
		// },

		// Test Case 3: Bad request (invalid value for gauge)
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
