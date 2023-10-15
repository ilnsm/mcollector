package agent

import (
	"errors"
	"fmt"
	"github.com/ilnsm/mcollector/internal/agent/config"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const pollCounter = 1
const deafaultSchema = "http://"
const updatePath = "/update"

func Run() {

	cfg, err := config.New()
	if err != nil {
		log.Fatal("Could not get config")
	}

	fmt.Printf("Start server\nPush to %s\nCollecting metrigs every %v\n"+
		"Send metrics every %v\n", cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)
	m := runtime.MemStats{}
	client := &http.Client{}

	for {

		metrics, err := GetMetrics(&m, cfg.PollInterval)
		if err != nil {
			fmt.Println("could not get metrics")
		}

		for name, value := range metrics {

			err := makeReq(cfg.Endpoint, "gauge", name, value, client)
			if err != nil {
				fmt.Println("could create request")
			}

		}

		err = makeReq(cfg.Endpoint, "counter", "PollCount", strconv.Itoa(pollCounter), client)
		if err != nil {
			fmt.Println("could create request")
		}

		randomFloat := rand.Float64()

		err = makeReq(cfg.Endpoint, "gauge", "RandomValue", strconv.FormatFloat(randomFloat, 'f', -1, 64), client)
		if err != nil {
			fmt.Println("could create request")
		}

		time.Sleep(cfg.ReportInterval)
	}
}

func doRequest(request *http.Request, client *http.Client) {
	request.Header.Add("Content-Type", "text/plain")
	r, err := client.Do(request)
	if err != nil {
		fmt.Printf("could not do request: %s", request.RequestURI)
	}
	err = r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func makeReq(endpoint, mtype, name, value string, client *http.Client) error {
	endpoint = deafaultSchema + endpoint + updatePath
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s/%s", endpoint, mtype, name, value), nil)
	if err != nil {
		return errors.New("could create request")
	}
	doRequest(request, client)
	return nil
}
