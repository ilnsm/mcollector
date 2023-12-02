package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

const compressFunc = "gzip"

var allowedContentTypes = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func DecompressRequest(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const wrapErr = "middleware compressor"

			// If content type does not match with allowedContentTypes stop processing and return to next handler
			if !matchContentTypes(r.Header.Values("Content-Type"), allowedContentTypes) {
				log.Debug().Msgf("did not match Content-Type in DecompressRequest")
				next.ServeHTTP(w, r)
				return
			}
			log.Debug().Msgf("matched Content-Type in DecompressRequest")

			// If compress function does not match with compressFunc stop processing and return to next handler
			if !matchCompressFunc(r.Header.Values("Content-Encoding"), compressFunc) {
				log.Debug().Msg("did not match Content-Encoding")
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error().Err(err).Msg(wrapErr)
			}
			decompressed, err := decompressGzip(body, log)
			if err != nil {
				log.Error().Err(err).Msg(wrapErr)
				http.Error(w, "failed to decompress data", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(strings.NewReader(string(decompressed)))

			next.ServeHTTP(w, r)
		})
	}
}

func CompressResponse(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const wrapError = "middleware compressor"

			// If client does not support compressed body stop processing and return to next handler
			if !matchCompressFunc(r.Header.Values("Accept-Encoding"), compressFunc) {
				log.Debug().Msg("did not match  Accept-Encoding")
				next.ServeHTTP(w, r)
				return
			}
			log.Debug().Msg("match Accept-Encoding")
			gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
			if err != nil {
				log.Error().Err(err).Msg(wrapError)
			}
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")

			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
		})
	}
}

func matchContentTypes(headers []string, allowedTypes []string) bool {
	if len(headers) == 0 {
		return false
	}
	for _, curType := range headers {
		for _, allowType := range allowedTypes {
			if curType == allowType {
				return true
			}
		}
	}
	return false
}

func matchCompressFunc(headers []string, compressFunc string) bool {
	if len(headers) == 0 {
		return false
	}
	for _, h := range headers {
		if h == compressFunc {
			return true
		}
	}
	return false
}

func decompressGzip(data []byte, log zerolog.Logger) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompressGzip error: %w", err)
	}
	func() {
		err := r.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close gzip.NewReader")
		}
	}()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("decompressGzip error: %w", err)
	}
	return b.Bytes(), nil
}
