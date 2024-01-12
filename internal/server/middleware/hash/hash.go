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

const hashHeader = "HashSHA256"

func VerifyRequestBodyIntegrity(log zerolog.Logger, key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Str("middleware", "VerifyRequestBodyIntegrity").Logger()

			hash := r.Header.Get(hashHeader)
			if hash == "" {
				l.Debug().Msg("request has no hash header")
				next.ServeHTTP(w, r)
				return
			}

			l.Debug().Msgf("got hash header %s", hash)

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
