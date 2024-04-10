// Package ssl provides functions for handling SSL encryption in HTTP requests.
package ssl

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/rs/zerolog"
)

// ParsePrivateKey reads a PEM-encoded private key from a file, decodes it,
// parses it into an ECDSA private key, and then imports it into an ECIES private key.
func ParsePrivateKey(path string) (*ecies.PrivateKey, error) {
	// Read the private key from file
	privKeyPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read PEM file: %w", err)
	}

	// Decode the PEM-encoded private key
	block, _ := pem.Decode(privKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse private key PEM: %w", err)
	}

	// Parse the private key
	privatekey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Import the ECDSA private key to an ECIES private key
	privatekeyECIES := ecies.ImportECDSA(privatekey)

	return privatekeyECIES, nil
}

// Terminate is a middleware function that decrypts the body of incoming HTTP requests.
// It reads the body of the request, decrypts it using the provided private key,
// and then replaces the original body with the decrypted data.
func Terminate(log zerolog.Logger, privkey *ecies.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := log.With().Str("middleware", "TerminateSSL").Logger()

			cyphertext, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				l.Debug().Msg("failed to read the body")
				return
			}

			plaintext, err := privkey.Decrypt(cyphertext, nil, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				l.Debug().Msg("failed to decrypt the body")
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(plaintext))
			next.ServeHTTP(w, r)
		})
	}
}
