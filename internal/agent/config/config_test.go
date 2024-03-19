package config

import (
	"testing"
	"time"

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

	c, err := New()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:9090", c.Endpoint)
	assert.Equal(t, "testkey", c.Key)
	assert.Equal(t, "debug", c.LogLevel)
	assert.Equal(t, time.Duration(60)*time.Second, c.ReportInterval)
	assert.Equal(t, time.Duration(30)*time.Second, c.PollInterval)
	assert.Equal(t, 100, c.RateLimit)
}

func TestNewConfigWithInvalidEnvironmentVariables(t *testing.T) {
	t.Setenv("REPORT_INTERVAL", "invalid")
	t.Setenv("POLL_INTERVAL", "invalid")
	t.Setenv("RATE_LIMIT", "invalid")
	_, err := New()
	assert.Error(t, err)
}
