package server

import (
	"fmt"
	"github.com/ilnsm/mcollector/internal/storage"
	memoryStorage "github.com/ilnsm/mcollector/internal/storage/memory"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Run() error {

	ms, err := memoryStorage.New()
	if err != nil {
		log.Fatal("could not inizialize storage")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/update/gauge/", updateGauge(ms))
	mux.HandleFunc("/update/counter/", updateCaunter(ms))
	mux.HandleFunc("/", handleBadRequest)
	return http.ListenAndServe("localhost:8080", mux)
}

func handleBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}

func updateGauge(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		//metric has no value
		if len(parts) < 5 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		//metric has no name
		metricName, metricValue := parts[3], parts[4]

		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			fmt.Println("error convert string to int64")
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
		err = s.InsertGauge(metricName, v)
		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}

func updateCaunter(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		//metric has no value
		if len(parts) < 5 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		//metric has no name
		metricName, metricValue := parts[3], parts[4]

		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			fmt.Println("error convert string to int64")
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
		err = s.InsertCounter(metricName, v)

		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}
