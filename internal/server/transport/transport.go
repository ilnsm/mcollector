package transport

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ilnsm/mcollector/internal/models"
	"github.com/rs/zerolog/log"
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

func UpdateTheMetric(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType, mName, mValue := chi.URLParam(r, "mType"), chi.URLParam(r, "mName"), chi.URLParam(r, "mValue")
		switch mType {
		case "gauge":
			{
				v, err := strconv.ParseFloat(mValue, 64)
				if err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
				}
				err = a.Storage.InsertGauge(mName, v)
				if err != nil {
					http.Error(w, "Not Found", http.StatusBadRequest)
				}

				w.WriteHeader(http.StatusOK)
			}

		case "counter":
			{
				v, err := strconv.ParseInt(mValue, 10, 64)
				if err != nil {
					http.Error(w, "Bad request", http.StatusBadRequest)
				}
				err = a.Storage.InsertCounter(mName, v)

				if err != nil {
					http.Error(w, "Not Found", http.StatusBadRequest)
				}

				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func GetTheMetric(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType, mName := chi.URLParam(r, "mType"), chi.URLParam(r, "mName")

		switch mType {
		case "gauge":
			v, err := a.Storage.SelectGauge(mName)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			_, err = w.Write([]byte(strconv.FormatFloat(v, 'g', -1, 64)))
			if err != nil {
				log.Err(err)
			}

		case "counter":
			{
				v, err := a.Storage.SelectCounter(mName)
				if err != nil {
					http.NotFound(w, r)
					return
				}
				_, err = w.Write([]byte(strconv.FormatInt(v, 10)))
				if err != nil {
					log.Err(err)
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func ListAllMetrics(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, g := a.Storage.GetCounters(), a.Storage.GetGauges()
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

func UpdateTheMetricWithJSON(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case "gauge":
			err := a.Storage.InsertGauge(m.ID, *m.Value)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			*m.Value, err = a.Storage.SelectGauge(m.ID)
			if err != nil {
				a.Log.Error().Msg("error get gauge's value ")
			}
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("connote encode answer")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("sending HTTP 200 response")

		case "counter":
			err := a.Storage.InsertCounter(m.ID, *m.Delta)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			*m.Delta, err = a.Storage.SelectCounter(m.ID)
			if err != nil {
				a.Log.Error().Msg("error get counter's value ")
			}
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("connote encode answer")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("sending HTTP 200 response")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func GetTheMetricWithJSON(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case "gauge":
			value, err := a.Storage.SelectGauge(m.ID)
			if err != nil {
				a.Log.Error().Msg("error getting gauge's value")
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")
				return
			}
			m.Value = &value
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("connote encode answer")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("sending HTTP 200 response")

		case "counter":
			delta, err := a.Storage.SelectCounter(m.ID)
			if err != nil {
				a.Log.Error().Msg("error getting counter's value")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			m.Delta = &delta
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("connote encode answer")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("sending HTTP 200 response")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}
