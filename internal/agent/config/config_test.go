package config

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigWithDefaultValues(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:8080", c.Endpoint)
	assert.Equal(t, "", c.Key)
	assert.Equal(t, "error", c.LogLevel)
	assert.Equal(t, time.Duration(defaultReportInterval)*time.Second, c.ReportInterval)
	assert.Equal(t, time.Duration(defaultPollInterval)*time.Second, c.PollInterval)
	assert.Equal(t, 1, c.RateLimit)
}

func TestNewConfigWithEnvironmentVariables(t *testing.T) {
	t.Setenv("ADDRESS", "localhost:9090")
	t.Setenv("KEY", "testkey")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REPORT_INTERVAL", "60")
	t.Setenv("POLL_INTERVAL", "30")
	t.Setenv("RATE_LIMIT", "100")
	t.Setenv("CRYPTO_KEY", "testkey")

	c, err := New()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:9090", c.Endpoint)
	assert.Equal(t, "testkey", c.Key)
	assert.Equal(t, "debug", c.LogLevel)
	assert.Equal(t, time.Duration(60)*time.Second, c.ReportInterval)
	assert.Equal(t, time.Duration(30)*time.Second, c.PollInterval)
	assert.Equal(t, 100, c.RateLimit)
	assert.Equal(t, "testkey", c.CryptoKey)
}

func TestNewConfigWithInvalidEnvironmentVariables(t *testing.T) {
	t.Setenv("REPORT_INTERVAL", "invalid")
	t.Setenv("POLL_INTERVAL", "invalid")
	t.Setenv("RATE_LIMIT", "invalid")
	_, err := New()
	assert.Error(t, err)
}

func TestConfigFileWithEnvironmentVariables(t *testing.T) {
	// Set environment variable
	t.Setenv("ADDRESS", "localhost:9090")
	t.Setenv("REPORT_INTERVAL", "333")
	t.Setenv("POLL_INTERVAL", "335")
	t.Setenv("CRYPTO_KEY", "/crypto/foo_crypto")

	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config.*.json")
	if err != nil {
		log.Fatal().Err(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Write a config file with different address
	text := `{"address": "localhost:8080", 
			  "report_interval": "1s",
			  "poll_interval": "1s",
			  "crypto_key": "/path/to/key.pem"}`
	if _, err := tmpfile.Write([]byte(text)); err != nil {
		log.Fatal().Err(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal().Err(err)
	}

	// Set the CONFIG environment variable to the temp file path
	t.Setenv("CONFIG", tmpfile.Name())

	// Call the New function to get the config
	c, err := New()
	assert.NoError(t, err)

	// Check that the address from the environment variable is used, not the one from the config file
	assert.Equal(t, "localhost:9090", c.Endpoint)
	assert.Equal(t, 333*time.Second, c.ReportInterval)
	assert.Equal(t, 335*time.Second, c.PollInterval)
	assert.Equal(t, "/crypto/foo_crypto", c.CryptoKey)
}
