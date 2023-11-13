package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			log.Info().
				Str("URI", uri).
				Str("Method", method).
				Str("Duration", time.Since(start).String()).
				Int("Bytes", ww.BytesWritten()).
				Int("Status", ww.Status()).
				Msg("")
		}()

		next.ServeHTTP(ww, r)
	})
}
