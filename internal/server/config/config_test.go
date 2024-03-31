package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ospiem/mcollector/internal/server/config"
)

func TestNew(t *testing.T) {
	t.Run("returns default config when no environment variables are set", func(t *testing.T) {
		c, err := config.New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:8080", c.Endpoint)
		assert.Equal(t, "error", c.LogLevel)
		assert.Equal(t, "/tmp/metrics-db.json", c.StoreConfig.FileStoragePath)
		assert.Equal(t, true, c.StoreConfig.Restore)
		assert.Equal(t, "", c.StoreConfig.DatabaseDsn)
		assert.Equal(t, "", c.Key)
	})

	t.Run("returns updated config when environment variables are set", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:9090")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("KEY", "testkey")
		t.Setenv("CRYPTO_KEY", "testkey")

		c, err := config.New()
		assert.NoError(t, err)
		assert.Equal(t, "localhost:9090", c.Endpoint)
		assert.Equal(t, "debug", c.LogLevel)
		assert.Equal(t, "testkey", c.Key)
		assert.Equal(t, "testkey", c.CryptoKey)
	})
}
