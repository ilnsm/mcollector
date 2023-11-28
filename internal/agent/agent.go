package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

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

	mTicker := time.NewTicker(cfg.PollInterval)
	reqTicker := time.NewTicker(cfg.ReportInterval)

	var pollCounter int
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
				err := makeReq(cfg.Endpoint, gauge, name, value, client)
				if err != nil {
					log.Err(err)
				}
			}

			err = makeReq(cfg.Endpoint, counter, "PollCount", strconv.Itoa(pollCounter), client)
			if err != nil {
				log.Err(err)
			}

			randomFloat := rand.Float64()

			err = makeReq(cfg.Endpoint, gauge, "RandomValue", strconv.FormatFloat(randomFloat, 'f', -1, 64), client)
			if err != nil {
				log.Err(err)
			}

			pollCounter = 0
		}
	}
}

func makeReq(endpoint, mtype, name, value string, client *http.Client) error {
	const wrapError = "make request error"
	endpoint = fmt.Sprintf("%v%v%v", defaultSchema, endpoint, updatePath)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s/%s", endpoint, mtype, name, value), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	request.Header.Add("Content-Type", "text/plain")
	err = doRequest(request, client)
	if err != nil {
		return err
	}
	return nil
}

func doRequest(request *http.Request, client *http.Client) error {
	const wrapError = "do request error"
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
