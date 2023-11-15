package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/storage"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>

	<title>Metric's' Data</title>

</head>
<body>

	   <h1>Data</h1>
	   <ul>
	   {{range $key, $value := .}}
	       <li>{{ $key }}: {{ $value }}</li>
	   {{end}}
	   </ul>


</body>
</html>
`

func CheckMetricType(next http.Handler) http.Handler {
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

func UpdateGauge(s storage.Storager) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		gName, gValue := chi.URLParam(r, "gName"), chi.URLParam(r, "gValue")

		v, err := strconv.ParseFloat(gValue, 64)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		err = s.InsertGauge(gName, v)
		if err != nil {
			http.Error(w, "Not Found", http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateCounter(s storage.Storager) http.HandlerFunc {
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

		w.WriteHeader(http.StatusOK)
	}
}

func GetGauge(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		k := chi.URLParam(r, "gName")
		v, err := s.SelectGauge(k)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatFloat(v, 'g', -1, 64)))
		if err != nil {
			log.Err(err)
		}

	}
}
func GetCounter(s storage.Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		k := chi.URLParam(r, "cName")
		v, err := s.SelectCounter(k)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		_, err = w.Write([]byte(strconv.FormatInt(v, 10)))
		if err != nil {
			log.Err(err)
		}

	}
}

func ListAllMetrics(s storage.Storager) http.HandlerFunc {
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
