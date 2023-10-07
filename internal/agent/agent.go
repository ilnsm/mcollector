package agent

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

const endpoint = "http://localhost:8080"

func Run() {

	m := runtime.MemStats{}
	client := &http.Client{}

	for {

		metrics, err := GetMetrics(&m)
		if err != nil {
			fmt.Println("could not get metrics")
		}

		for name, value := range metrics {

			request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/gauge/%s/%s", endpoint, name, value), nil)
			if err != nil {
				fmt.Println("could not read metric")
			}

			request.Header.Add("Content-Type", "text/plain")
			_, err = client.Do(request)
			if err != nil {
				fmt.Println("could not read metric")
			}
		}

		time.Sleep(time.Second * 2)
	}
}
