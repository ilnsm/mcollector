package logger

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		next.ServeHTTP(w, r)
		duration := time.Since(start)

		log.Info().
			Str("URI", uri).
			Str("Method", method).
			Str("Duration", duration.String()).
	})
}
