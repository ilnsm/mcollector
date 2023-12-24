package transport

import (
	"context"
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
const contentType = "Content-Type"
const applicationJSON = "application/json"
const internalServerError = "Internal server error"

func UpdateTheMetric(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType, mName, mValue := chi.URLParam(r, "mType"), chi.URLParam(r, "mName"), chi.URLParam(r, "mValue")
		switch mType {
		case models.Gauge:
			{
				v, err := strconv.ParseFloat(mValue, 64)
				if err != nil {
					http.Error(w, "Bad request to update gauge", http.StatusBadRequest)
				}
				err = a.Storage.InsertGauge(ctx, mName, v)
				if err != nil {
					http.Error(w, "Not Found", http.StatusBadRequest)
				}

				w.WriteHeader(http.StatusOK)
			}

		case models.Counter:
			{
				v, err := strconv.ParseInt(mValue, 10, 64)
				if err != nil {
					http.Error(w, "Bad request to update counter", http.StatusBadRequest)
				}
				err = a.Storage.InsertCounter(ctx, mName, v)

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

func GetTheMetric(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType, mName := chi.URLParam(r, "mType"), chi.URLParam(r, "mName")

		switch mType {
		case models.Gauge:
			v, err := a.Storage.SelectGauge(ctx, mName)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			_, err = w.Write([]byte(strconv.FormatFloat(v, 'g', -1, 64)))
			if err != nil {
				log.Err(err)
			}

		case models.Counter:
			{
				v, err := a.Storage.SelectCounter(ctx, mName)
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

func ListAllMetrics(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, g := a.Storage.GetCounters(ctx), a.Storage.GetGauges(ctx)
		var data = make(map[string]string)
		for i, v := range c {
			data[i] = strconv.Itoa(int(v))
		}
		for i, v := range g {
			data[i] = strconv.FormatFloat(v, 'f', -1, 64)
		}
		w.Header().Set(contentType, "text/html")
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func UpdateTheMetricWithJSON(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case models.Gauge:
			err := a.Storage.InsertGauge(ctx, m.ID, *m.Value)
			if err != nil {
				http.Error(w, internalServerError, http.StatusInternalServerError)
				return
			}

			*m.Value, err = a.Storage.SelectGauge(ctx, m.ID)
			if err != nil {
				a.Log.Error().Msg("error get gauge's value ")
			}
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("UpdateTheMetricWithJSON: sending HTTP 200 response")

		case models.Counter:
			err := a.Storage.InsertCounter(ctx, m.ID, *m.Delta)
			if err != nil {
				http.Error(w, internalServerError, http.StatusInternalServerError)
				return
			}

			*m.Delta, err = a.Storage.SelectCounter(ctx, m.ID)
			if err != nil {
				a.Log.Error().Msg("error get counter's value ")
			}
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "UpdateTheMetricWithJSON").Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("UpdateTheMetricWithJSON: sending HTTP 200 response")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func GetTheMetricWithJSON(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case models.Gauge:
			value, err := a.Storage.SelectGauge(ctx, m.ID)
			if err != nil {
				a.Log.Error().Msg("error getting gauge's value")
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set(contentType, applicationJSON)
				return
			}
			m.Value = &value
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "GetTheMetricWithJSON").Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("GetTheMetricWithJSON: sending HTTP 200 response")

		case models.Counter:
			delta, err := a.Storage.SelectCounter(ctx, m.ID)
			if err != nil {
				a.Log.Error().Msg("error getting counter's value")
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set(contentType, applicationJSON)
				return
			}
			m.Delta = &delta
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				a.Log.Error().Str("func", "GetTheMetricWithJSON").Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			a.Log.Debug().Msg("GetTheMetricWithJSON: sending HTTP 200 response")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func UpdateSliceOfMetrics(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := a.Storage.InsertBatch(ctx, metrics); err != nil {
			a.Log.Error().Err(err).Msg("cannot insert batch in handler")
		}
		w.WriteHeader(http.StatusOK)
		a.Log.Debug().Msg("UpdateSliceOfMetrics: sending HTTP 200 response")
	}
}

func PingDB(ctx context.Context, a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Storage.Ping(ctx); err != nil {
			http.Error(w, internalServerError, http.StatusInternalServerError)
		}
	}
}
