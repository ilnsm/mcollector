package hash

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestVerifyRequestBodyIntegrity(t *testing.T) {
	type testCases struct {
		name     string
		body     []byte
		hash     string
		key      string
		expected bool
	}

	tests := []testCases{
		{
			name: "Negative Test",
			body: []byte(`[{"id":"gaugeNameBatchJSON","type":"gauge","value":138.988},
  					{"id":"counterNameBatchJSON","type":"counter","delta":113}]`),
			hash:     "810f6b94525bcfac28f9870a314fdc151bdb6427bd0acdb6574100a98643945a",
			key:      "fiok120uo8i3rhfkw",
			expected: false,
		},

		{
			name: "Positive Test",
			body: []byte(`[{"id":"gaugeNameBatchJSON","type":"gauge","value":138.988},
  					{"id":"counterNameBatchJSON","type":"counter","delta":113}]`),
			hash:     "7d2fe9de5dd00a43ccff6d7f0c85e4fbc7ba19a903bfbe674c57a739523a10c4",
			key:      "fiok120uo8i3rhfkw",
			expected: true,
		},
	}

	l := zerolog.Logger{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/test", nil)
			request.Header.Set(hashHeader, tc.hash)
			request.Body = io.NopCloser(bytes.NewBuffer(tc.body))

			handler := VerifyRequestBodyIntegrity(l, tc.key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, request)

			if tc.expected {
				if w.Code != http.StatusOK {
					t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected status code %d, got %d", http.StatusBadRequest, w.Code)
				}
			}
		})
	}
}
