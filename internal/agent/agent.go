// Package agent provides functionality for managing metrics collection and reporting.

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ospiem/mcollector/internal/agent/config"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/rs/zerolog"
)

// defaultSchema defines the default schema for HTTP requests.
const defaultSchema = "http://"

// updatePath defines the path for updating metrics.
const updatePath = "/updates/"

// gauge represents the metric type "gauge".
const gauge = "gauge"

// counter represents the metric type "counter".
const counter = "counter"

// cannotCreateRequest is the error message when a request cannot be created.
const cannotCreateRequest = "cannot create request"

// retryAttempts specifies the number of retry attempts.
const retryAttempts = 3

// repeatFactor specifies the factor for increasing sleep time between retries.
const repeatFactor = 2

// errRetryableHTTPStatusCode is the error for retryable HTTP status codes.
var errRetryableHTTPStatusCode = errors.New("got retryable status code")

// MetricsCollection represents a collection of metrics.
type MetricsCollection struct {
	mux  *sync.Mutex
	coll map[string]string
}

// NewMetricsCollection creates a new MetricsCollection instance.
func NewMetricsCollection() *MetricsCollection {
	return &MetricsCollection{
		coll: make(map[string]string),
		mux:  &sync.Mutex{},
	}
}

// Push adds metrics to the collection.
func (mc *MetricsCollection) Push(metrics map[string]string) {
	mc.mux.Lock()
	defer mc.mux.Unlock()
	mc.coll = metrics
}

// Pop retrieves metrics from the collection.
func (mc *MetricsCollection) Pop() map[string]string {
	mc.mux.Lock()
	defer mc.mux.Unlock()
	return mc.coll
}

// Worker represents a worker that processes metrics.
func Worker(ctx context.Context, wg *sync.WaitGroup, cfg config.Config,
	dataChan chan map[string]string, log zerolog.Logger) {
	defer wg.Done()
	l := log.With().Str("func", "worker").Logger()
	reqTicker := time.NewTicker(cfg.ReportInterval)
	defer reqTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("Stopping worker")
			return
		case <-reqTicker.C:

			metrics := <-dataChan
			var pollIncrement int64 = 1
			client := &http.Client{}
			var metricSlice []models.Metrics

			for name, value := range metrics {
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					l.Error().Err(err).Msg("error convert string to float")
					continue
				}
				metricSlice = append(metricSlice, models.Metrics{MType: gauge, ID: name, Value: &v})
			}

			randomFloat := rand.Float64()
			metricSlice = append(metricSlice, models.Metrics{MType: gauge, ID: "RandomValue", Value: &randomFloat},
				models.Metrics{MType: counter, ID: "PollCount", Delta: &pollIncrement})

			attempt := 0
			sleepTime := 1 * time.Second

			for {
				select {
				case <-ctx.Done():
					l.Info().Msg("Stopping worker")
					return
				default:
				}

				var opError *net.OpError
				l.Debug().Msg("Trying to send request")
				err := doRequestWithJSON(cfg, metricSlice, client, log)
				if err == nil {
					break
				}
				if errors.As(err, &opError) || errors.Is(err, errRetryableHTTPStatusCode) {
					l.Error().Err(err).Msgf("%s, will retry in %v", cannotCreateRequest, sleepTime)
					time.Sleep(sleepTime)
					attempt++
					sleepTime += repeatFactor * time.Second
					if attempt < retryAttempts {
						continue
					}
					break
				}
				l.Error().Err(err).Msgf("cannot do request, failed %d times", retryAttempts)
			}
		}
	}
}

// doRequestWithJSON sends a request with JSON data.
func doRequestWithJSON(cfg config.Config, metrics []models.Metrics, client *http.Client, l zerolog.Logger) error {
	const wrapError = "do request error"

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err = g.Write(jsonData); err != nil {
		return fmt.Errorf("create gzip in %s: %w", wrapError, err)
	}
	if err = g.Close(); err != nil {
		return fmt.Errorf("close gzip in %s: %w", wrapError, err)
	}

	ep := fmt.Sprintf("%v%v%v", defaultSchema, cfg.Endpoint, updatePath)

	request, err := http.NewRequest(http.MethodPost, ep, &buf)
	if err != nil {
		return fmt.Errorf("generate request %s: %w", wrapError, err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	if cfg.Key != "" {
		request.Header.Set("HashSHA256", generateHash(cfg.Key, jsonData, l))
	}

	r, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("body close %s: %w", wrapError, err)
	}

	if isStatusCodeRetryable(r.StatusCode) {
		return errRetryableHTTPStatusCode
	}

	return nil
}

// isStatusCodeRetryable checks if a status code is retryable.
func isStatusCodeRetryable(code int) bool {
	switch code {
	case
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// generateHash generates a hash for the given key and data.
func generateHash(key string, data []byte, l zerolog.Logger) string {
	logger := l.With().Str("func", "generateHash").Logger()
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(data)
	if err != nil {
		logger.Error().Err(err).Msg("cannot hash data")
		return ""
	}

	hash := hex.EncodeToString(h.Sum(nil))
	logger.Debug().Msg(hash)

	return hash
}
