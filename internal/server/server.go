package server

import (
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

var ms = MemStorage{gauge: make(map[string]float64), counter: make(map[string]int64)}

func Run() error {

	mux := http.NewServeMux()

	mux.HandleFunc("/update/", updateMetrics)

	return http.ListenAndServe("localhost:8080", mux)
}

func updateMetrics(w http.ResponseWriter, r *http.Request) {

	//if ct := r.Header.Get("Content-Type"); ct != "text/plain" {
	//	http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
	//}

	parts := strings.Split(r.URL.Path, "/")

	//metric has no value
	if len(parts) < 5 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	//metric has no name
	metricName := parts[3]
	if metricName == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	switch parts[2] {
	case "counter":
		v, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		ms.counter[parts[3]] += v
	case "gauge":
		v, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		ms.gauge[parts[3]] = v
	default:
		http.Error(w, "Metric's type does not support", http.StatusBadRequest)
		return
	}
}
