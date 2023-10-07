package agent

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const endpoint = "http://localhost:8080/update"
const pollInterval = time.Second * 2
const reportInterval = time.Second * 10

func Run() {

	m := runtime.MemStats{}
	client := &http.Client{}
	var counter int

	for {

		metrics, err := GetMetrics(&m, pollInterval)
		if err != nil {
			fmt.Println("could not get metrics")
		}

		for name, value := range metrics {

			err := makeReq("gauge", name, value, client)
			if err != nil {
				fmt.Println("could create request")
			}

		}

		err = makeReq("counter", "PollCount", strconv.Itoa(counter), client)
		if err != nil {
			fmt.Println("could create request")
		}

		randomFloat := rand.Float64()

		err = makeReq("gauge", "RandomValue", strconv.FormatFloat(randomFloat, 'f', -1, 64), client)
		if err != nil {
			fmt.Println("could create request")
		}

		counter++
		time.Sleep(reportInterval)
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

func makeReq(mtype, name, value string, client *http.Client) error {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s/%s", endpoint, mtype, name, value), nil)
	if err != nil {
		return errors.New("could create request")
	}
	doRequest(request, client)
	return nil
}
