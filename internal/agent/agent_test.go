package agent

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ospiem/mcollector/internal/agent/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMetricsCollection_PushAndPop(t *testing.T) {
	mc := NewMetricsCollection()
	metrics := map[string]string{
		"metric1": "value1",
		"metric2": "value2",
	}

	mc.Push(metrics)
	result := mc.Pop()

	assert.Equal(t, metrics, result)
}

func TestIsStatusCodeRetryable(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{
			name: "Retryable status code",
			code: http.StatusInternalServerError,
			want: true,
		},
		{
			name: "Non-retryable status code",
			code: http.StatusOK,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStatusCodeRetryable(tt.code)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	cfg := config.Config{
		ReportInterval: time.Millisecond * 100,
		Endpoint:       "localhost:8080",
		Key:            "testKey",
	}

	dataChan := make(chan map[string]string, 1)
	dataChan <- map[string]string{
		"metric1": "value1",
		"metric2": "value2",
	}

	log := zerolog.Nop()

	go Worker(ctx, wg, cfg, dataChan, log)

	time.Sleep(time.Millisecond * 200)
	cancel()

	wg.Wait()
}

func TestGenerateHash(t *testing.T) {
	key := "fiok120uo8i3rhfkw"
	data := []byte("testData")
	expectedHash := "9d4a553aa9fb8670764fb8351062784369646e8f53ba9c2ef8e50bb241887c3b"

	result := generateHash(key, data, zerolog.Nop())

	assert.Equal(t, expectedHash, result)
}

func TestGetMetrics(t *testing.T) {
	metrics, err := GetMetrics()
	assert.NoError(t, err)

	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
		"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc",
	}

	for _, metric := range expectedMetrics {
		_, exists := metrics[metric]
		assert.True(t, exists, "Expected metric %s does not exist", metric)
	}
}

func TestEncryptDataWithInvalidKeyPath(t *testing.T) {
	_, err := encryptData([]byte("testData"), "nonexistent.pem")
	assert.Error(t, err)
}

func TestEncryptDataWithInvalidCertificate(t *testing.T) {
	err := os.WriteFile("invalid.pem", []byte("invalid"), 0644)
	assert.NoError(t, err)
	_, err = encryptData([]byte("testData"), "/tmp/invalid.pem")
	assert.Error(t, err)
	os.Remove("/tmp/invalid.pem")
}

func TestEncryptDataWithValidCertificate(t *testing.T) {
	err := os.WriteFile("/tmp/valid.pem", []byte(`-----BEGIN CERTIFICATE-----
MIIBUzCB2qADAgECAgEBMAoGCCqGSM49BAMCMBUxEzARBgNVBAoTCm1jb2xsZWN0
b3IwHhcNMjQwMzMxMTM1ODU1WhcNMzQwMzMxMTM1ODU1WjAVMRMwEQYDVQQKEwpt
Y29sbGVjdG9yMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEQ08QQSIFpW5S+sxDm1/4
/hG4UJrPd3SY4m/MN0PKdrscZncrzS6cmiJ0JErxOle06bQSRRA/CgIV6qPDKtS4
thJEFEqLzIsr+3SJvDmX4xGutdJQmcj3AQSlS2R38CBsMAoGCCqGSM49BAMCA2gA
MGUCMQCNAqIjkUlhQUuyaKOuO2gJbr92lxIL5tYkIJ6johEi4aRjCLPOLKf2Lnb4
IoZJD6oCMDuhQlLu3fV4BLuSiHIXGp56mHG9FpWdFvNq5i7g3bkxt4bbwMdLCeyf
t0IlJDQqiw==
-----END CERTIFICATE-----
`), 0600)
	assert.NoError(t, err)
	_, err = encryptData([]byte("testData"), "/tmp/valid.pem")
	assert.NoError(t, err)
	os.Remove("/tmp/valid.pem")
}
