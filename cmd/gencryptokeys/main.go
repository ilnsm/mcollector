// Package main provides a command-line tool for generating ECDSA private and public keys.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const fileMode = 0600
const dirMode = 0700
const certificateValidityYears = 10
const certPem = "cert.pem"
const privKeyPem = "privkey.pem"

// main is the entry point of the application.
// It parses command-line arguments and calls the generateKeys function.
func main() {
	var path string
	flag.StringVar(&path, "p", "keys", "Path to save generated keys")
	flag.Parse()

	err := generateKeys(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Log().Msgf("Certificates have been generated successfully. You can find them in %s", path)
}

// generateKeys generates a pair of ECDSA private and public keys and saves them to the specified path.
// The keys are saved in PEM format. The public key is saved in a X.509 certificate.
func generateKeys(path string) error {
	// Generate a pair of keys
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("cannot generate private key: %w", err)
	}
	publicKey := privateKey.Public()

	// Define template for x509 certificate
	template := x509.Certificate{
		SerialNumber:       big.NewInt(1),
		Subject:            pkix.Name{Organization: []string{"mcollector"}},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(certificateValidityYears, 0, 0),
		SignatureAlgorithm: x509.ECDSAWithSHA256, // Use ECDSA with SHA256
		PublicKeyAlgorithm: x509.ECDSA,           // Use ECDSA for public key
	}
	// Create a certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("cannot create a certificate: %w", err)
	}

	// Write certificate to file
	err = os.Mkdir(path, dirMode)
	if err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	certPEM, err := os.Create(fmt.Sprintf("%s/%s", path, certPem))
	if err != nil {
		return fmt.Errorf("cannot create file cert.pem: %w", err)
	}
	defer func() {
		err := certPEM.Close()
		if err != nil {
			log.Err(err).Msg("Error while closing cert.pem")
		}
	}()

	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return fmt.Errorf("cannot encode certificate: %w", err)
	}

	// Write private key to file
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("cannot marshal private key: %w", err)
	}
	privateKeyPEM, err := os.Create(fmt.Sprintf("%s/%s", path, privKeyPem))
	if err != nil {
		return fmt.Errorf("cannot create file privkey.pem: %w", err)
	}
	defer func() {
		err := privateKeyPEM.Close()
		if err != nil {
			log.Err(err).Msg("Error while closing privkey.pem")
		}
	}()

	err = pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err != nil {
		return fmt.Errorf("cannot encode private key: %w", err)
	}

	err = os.Chmod(fmt.Sprintf("%s/%s", path, certPem), fileMode)
	if err != nil {
		return fmt.Errorf("cannot change file mode: %w", err)
	}
	err = os.Chmod(fmt.Sprintf("%s/%s", path, privKeyPem), fileMode)
	if err != nil {
		return fmt.Errorf("cannot change file mode: %w", err)
	}

	return nil
}
