package agent

import (
	"context"
	"net/http"
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
