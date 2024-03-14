package compress

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gzip "github.com/klauspost/compress/gzip"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCompressMiddleware(t *testing.T) {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	tests := []struct {
		name           string
		handler        func(log zerolog.Logger) func(next http.Handler) http.Handler
		request        func() *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Decompress test data",
			handler: DecompressRequest,
			request: func() *http.Request {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				if _, err := gz.Write([]byte("test data")); err != nil {
					t.Fatal(err)
				}
				if err := gz.Close(); err != nil {
					t.Fatal(err)
				}

				req := httptest.NewRequest(http.MethodPost, "/test", &buf)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Encoding", "gzip")
				return req
			},
			expectedBody:   "test data",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("test data"))
				if err != nil {
					t.Fatal(err)
				}
			}))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.request())
			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}()

			assert.Equal(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
