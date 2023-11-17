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

const pollCounter = 1
const defaultSchema = "http://"
const updatePath = "/update"

func Run() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Msg("Could not get config")
	}

	log.Info().Msgf("Start server\nPush to %s\nCollecting metrigs every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)
	m := runtime.MemStats{}
	client := &http.Client{}

	for {
		metrics, err := GetMetrics(&m, time.Duration(cfg.PollInterval)*time.Second)
		if err != nil {
			log.Err(err)
		}

		for name, value := range metrics {
			err := makeReq(cfg.Endpoint, "gauge", name, value, client)
			if err != nil {
				log.Err(err)
			}
		}

		err = makeReq(cfg.Endpoint, "counter", "PollCount", strconv.Itoa(pollCounter), client)
		if err != nil {
			log.Err(err)
		}

		randomFloat := rand.Float64()

		err = makeReq(cfg.Endpoint, "gauge", "RandomValue", strconv.FormatFloat(randomFloat, 'f', -1, 64), client)
		if err != nil {
			log.Err(err)
		}

		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
	}
}

func doRequest(request *http.Request, client *http.Client) error {
	request.Header.Add("Content-Type", "text/plain")
	r, err := client.Do(request)
	if err != nil {
		return err
	}
	err = r.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func makeReq(endpoint, mtype, name, value string, client *http.Client) error {
	endpoint = defaultSchema + endpoint + updatePath
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s/%s", endpoint, mtype, name, value), nil)
	if err != nil {
		return err
	}
	err = doRequest(request, client)
	if err != nil {
		return err
	}
	return nil
}
