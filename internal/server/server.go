package server

import (
	"errors"
	"fmt"
	"github.com/ilnsm/mcollector/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

func Run(s storage.Storager) error {

	mux := http.NewServeMux()

	mux.HandleFunc("/update/gauge/", updateGauge(s))
	mux.HandleFunc("/update/counter/", updateCounter(s))
	mux.HandleFunc("/", handleBadRequest)
	return http.ListenAndServe("localhost:8080", mux)
}

func handleBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func updateGauge(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		err := mustHaveNameAndValue(parts)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		metricName, metricValue := parts[3], parts[4]

		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			fmt.Println("error convert string to int64")
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		err = s.InsertGauge(metricName, v)
		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}

func updateCounter(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		err := mustHaveNameAndValue(parts)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		metricName, metricValue := parts[3], parts[4]

		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		err = s.InsertCounter(metricName, v)

		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}

func mustHaveNameAndValue(p []string) error {
	if len(p) < 5 {
		return errors.New("mertric has no name or value")
	}
	return nil
}
