package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/ilnsm/mcollector/internal/models"

	"github.com/ilnsm/mcollector/internal/agent/config"
	"github.com/rs/zerolog/log"
)

const defaultSchema = "http://"
const updatePath = "/updates/"
const gauge = "gauge"
const counter = "counter"
const cannotCreateRequest = "cannot create request"
const retryAttempts = 3
const repeatFactor = 2

var errRetryableHTTPStatusCode = errors.New("got retryable status code")

func Run() error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("run agent error: %w", err)
	}

	log.Info().Msgf("Start server\nPush to %s\nCollecting metrics every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)

	m := runtime.MemStats{}
	metrics := make(map[string]string)
	client := &http.Client{}
	var metricSlice []models.Metrics

	mTicker := time.NewTicker(cfg.PollInterval)
	defer mTicker.Stop()
	reqTicker := time.NewTicker(cfg.ReportInterval)
	defer reqTicker.Stop()

	var pollCounter int64
	for {
		select {
		case <-mTicker.C:
			err := GetMetrics(&m, metrics)
			if err != nil {
				log.Err(err)
			}
			pollCounter++
		case <-reqTicker.C:
			for name, value := range metrics {
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Error().Msg("error convert string to float")
					break
				}
				metricSlice = append(metricSlice, models.Metrics{MType: gauge, ID: name, Value: &v})
			}
			randomFloat := rand.Float64()
			metricSlice = append(metricSlice, models.Metrics{MType: gauge, ID: "RandomValue", Value: &randomFloat},
				models.Metrics{MType: counter, ID: "PollCount", Delta: &pollCounter})

			attempt := 0
			sleepTime := 1 * time.Second
			for {
				var opError *net.OpError
				err = doRequestWithJSON(cfg.Endpoint, metricSlice, client)
				if err == nil {
					break
				}
				if errors.As(err, &opError) || errors.Is(err, errRetryableHTTPStatusCode) {
					log.Error().Err(err).Msgf("%s, will retry in %v", cannotCreateRequest, sleepTime)
					time.Sleep(sleepTime)
					attempt++
					sleepTime += repeatFactor * time.Second
					if attempt < retryAttempts {
						continue
					}
					break
				}
				log.Error().Err(err).Msgf("cannot do request, failed %d times", retryAttempts)
			}
			metricSlice = nil
			pollCounter = 0
		}
	}
}

func doRequestWithJSON(endpoint string, metrics []models.Metrics, client *http.Client) error {
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

	endpoint = fmt.Sprintf("%v%v%v", defaultSchema, endpoint, updatePath)

	request, err := http.NewRequest(http.MethodPost, endpoint, &buf)
	if err != nil {
		return fmt.Errorf("generate request %s: %w", wrapError, err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")

	r, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("body close %s: %w", wrapError, err)
	}

	if isStatusCoderetryable(r.StatusCode) {
		return errRetryableHTTPStatusCode
	}

	return nil
}

func isStatusCoderetryable(code int) bool {
	switch code {
	case http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
