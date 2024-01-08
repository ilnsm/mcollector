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
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ospiem/mcollector/internal/agent/config"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/ospiem/mcollector/internal/tools"
	"github.com/rs/zerolog"
)

const defaultSchema = "http://"
const updatePath = "/updates/"
const gauge = "gauge"
const counter = "counter"
const cannotCreateRequest = "cannot create request"
const retryAttempts = 3
const repeatFactor = 2
const workerPoolSizeFactor = 1

var errRetryableHTTPStatusCode = errors.New("got retryable status code")

func Run(ctx context.Context, wg *sync.WaitGroup) {
	cfg, err := config.New()
	if err != nil {
		return
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	tools.SetLogLevel(cfg.LogLevel)
	logger.Info().Msgf("Start server\nPush to %s\nCollecting metrics every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)

	ch := generator(ctx, cfg, logger)

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go worker(ctx, cfg, ch, logger)
	}
}

func generator(ctx context.Context, cfg config.Config, log zerolog.Logger) chan map[string]string {
	l := log.With().Str("func", "generator").Logger()
	mCHan := make(chan map[string]string, workerPoolSizeFactor*cfg.RateLimit)
	l.Debug().Msg("Hello from generator")
	go func() {
		defer close(mCHan)
		mTicker := time.NewTicker(cfg.PollInterval)
		defer mTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				l.Info().Msg("Stopping generator")
				return
			case <-mTicker.C:
				l.Debug().Msg("Trying to get metrics")
				m, err := GetMetrics()
				if err != nil {
					l.Error().Err(err).Msg("cannot get metrics")
					continue
				}
				mCHan <- m
			}
		}
	}()

	return mCHan
}

func worker(ctx context.Context, cfg config.Config, mCHan chan map[string]string, log zerolog.Logger) {
	l := log.With().Str("func", "worker").Logger()
	reqTicker := time.NewTicker(cfg.ReportInterval)
	defer reqTicker.Stop()

	l.Debug().Msg("Hello from worker")
	for {
		select {
		case <-ctx.Done():
			l.Info().Msg("Stopping worker")
			return
		default:
			var pollCounter int64
			for metrics := range mCHan {
				client := &http.Client{}
				var metricSlice []models.Metrics

				l.Debug().Msg("Trying to generate request")
				for name, value := range metrics {
					v, err := strconv.ParseFloat(value, 64)
					if err != nil {
						l.Error().Err(err).Msg("error convert string to float")
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

				metricSlice = nil
				pollCounter = 0
			}
		}
	}
}

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
