// Package hash provides middleware for verifying the integrity of the request body using HMAC-SHA256 hashing.
package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

// hashHeader is the HTTP header key for the expected hash.
const hashHeader = "HashSHA256"

// VerifyRequestBodyIntegrity returns a middleware that verifies
// the integrity of the request body using HMAC-SHA256 hashing.
// It compares the computed hash with the hash provided in the HTTP header.
// If the hashes match, the request proceeds to the next handler; otherwise, it returns a Bad Request error.
func VerifyRequestBodyIntegrity(log zerolog.Logger, key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Str("middleware", "VerifyRequestBodyIntegrity").Logger()

			hash := r.Header.Get(hashHeader)
			if hash == "" {
				next.ServeHTTP(w, r)
				return
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			r.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				if err := r.Body.Close(); err != nil {
					l.Error().Err(err)
				}
			}()

			h := hmac.New(sha256.New, []byte(key))
			_, err = h.Write(b)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
			}

			curHash := hex.EncodeToString(h.Sum(nil))
			if hash != curHash {
				http.Error(w, "Bad Request, hashes does not matched", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
