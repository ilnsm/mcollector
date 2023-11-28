package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/ilnsm/mcollector/internal/models"

	"github.com/ilnsm/mcollector/internal/agent/config"
	"github.com/rs/zerolog/log"
)

const defaultSchema = "http://"
const updatePath = "/update"
const gauge = "gauge"
const counter = "counter"

func Run() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Msg("Could not get config")
	}

	log.Info().Msgf("Start server\nPush to %s\nCollecting metrics every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)

	m := runtime.MemStats{}
	metrics := make(map[string]string)
	client := &http.Client{}
	var mModel models.Metrics

	mTicker := time.NewTicker(cfg.PollInterval)
	reqTicker := time.NewTicker(cfg.ReportInterval)

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
				mModel.ID = name
				mModel.MType = gauge
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Error().Msg("error convert string to float")
				}
				mModel.Value = &v

				err = doRequestWithJSON(cfg.Endpoint, mModel, client)
				if err != nil {
					log.Error().Err(err).Msg("cannot create request")
				}
			}

			mModel.ID = "PollCount"
			mModel.MType = counter
			mModel.Delta = &pollCounter
			err = doRequestWithJSON(cfg.Endpoint, mModel, client)
			if err != nil {
				log.Error().Err(err).Msg("cannot create request")
			}

			randomFloat := rand.Float64()
			mModel.ID = "RandomValue"
			mModel.MType = gauge
			mModel.Value = &randomFloat
			err = doRequestWithJSON(cfg.Endpoint, mModel, client)
			if err != nil {
				log.Error().Err(err).Msg("cannot create request")
			}
			pollCounter = 0
		}
	}
}

func doRequestWithJSON(endpoint string, m models.Metrics, client *http.Client) error {
	const wrapError = "do request error"
	endpoint = fmt.Sprintf("%v%v%v", defaultSchema, endpoint, updatePath)
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	request.Header.Set("Content-Type", "application/json")
	r, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	err = r.Body.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	return nil
}
