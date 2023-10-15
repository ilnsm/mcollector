package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/storage"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func MetrRouter(s storage.Storager) chi.Router {

	r := chi.NewRouter()
	r.Use(checkMetricType)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{gName}/{gValue}", updateGauge(s))
		r.Post("/counter/{cName}/{cValue}", updateCounter(s))
	})

	r.Get("/", listAllMetrics(s))

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{gName}", getGauge(s))
		r.Get("/counter/{cName}", getCounter(s))

	})
	return r
}

func checkMetricType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		p := r.URL.Path
		if len(p) > 2 {
			parts := strings.Split(p, "/")
			if parts[2] != "gauge" && parts[2] != "counter" {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func updateGauge(s storage.Storager) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		gName, gValue := chi.URLParam(r, "gName"), chi.URLParam(r, "gValue")

		v, err := strconv.ParseFloat(gValue, 64)
		if err != nil {
			//fmt.Println("error convert string to int64")
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		err = s.InsertGauge(gName, v)
		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}

func updateCounter(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cName, cValue := chi.URLParam(r, "cName"), chi.URLParam(r, "cValue")

		v, err := strconv.ParseInt(cValue, 10, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		err = s.InsertCounter(cName, v)

		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}
	}
}

func getGauge(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		k := chi.URLParam(r, "gName")
		v, err := s.SelectGauge(k)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatFloat(v, 'g', -1, 64)))

	}
}
func getCounter(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		k := chi.URLParam(r, "cName")
		v, err := s.SelectCounter(k)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(strconv.FormatInt(v, 10)))
	}
}

func listAllMetrics(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, g := s.GetCounters(), s.GetGauges()
		var data = make(map[string]string)
		for i, v := range c {
			data[i] = strconv.Itoa(int(v))
		}
		for i, v := range g {
			data[i] = strconv.FormatFloat(v, 'f', -1, 64)
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
