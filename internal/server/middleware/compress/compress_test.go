package compress

import (
	"bytes"
	"io"
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
		checkResponse  func(*testing.T, *http.Response)
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
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, "test data", string(body))
			},
		},
		{
			name:    "Compress test data",
			handler: CompressResponse,
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				return req
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				gr, err := gzip.NewReader(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				defer gr.Close()

				body, err := io.ReadAll(gr)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, "test data", string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.handler(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test data"))
			}))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, tt.request())

			assert.Equal(t, tt.expectedStatus, rr.Code)

			tt.checkResponse(t, rr.Result())
		})
	}
}
