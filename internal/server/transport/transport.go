package transport

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ospiem/mcollector/internal/models"
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
const sending200OK = "sending HTTP 200 response"
const invalidContentType = "invalid content-type"
const invalitContentTypeNotJSON = "Invalid Content-Type, expected application/json"

func UpdateTheMetric(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "UpdateTheMetric").Logger()
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
					logger.Error().Err(err).Msg("cannot parse counter")
					http.Error(w, "Bad request to update counter", http.StatusBadRequest)
				}
				err = a.Storage.InsertCounter(ctx, mName, v)

				if err != nil {
					logger.Error().Err(err).Msg("cannot update counter")
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
		ctx := r.Context()
		logger := a.Log.With().Str("func", "UpdateTheMetric").Logger()
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
					logger.Error().Err(err).Msg("cannot write response")
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func ListAllMetrics(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "ListAllMetrics").Logger()
		tmpl, err := template.New("index").Parse(htmlTemplate)
		if err != nil {
			logger.Error().Err(err).Msg("cannot create template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := a.Storage.GetCounters(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get counters")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g, err := a.Storage.GetGauges(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get gauges")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
			logger.Error().Err(err).Msg("cannot execute template")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func UpdateTheMetricWithJSON(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "UpdateTheMetricWithJSON").Logger()
		if r.Header.Get(contentType) != applicationJSON {
			http.Error(w, invalitContentTypeNotJSON, http.StatusBadRequest)
			logger.Debug().Msg(invalidContentType)
			return
		}
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case models.Gauge:
			err := a.Storage.InsertGauge(ctx, m.ID, *m.Value)
			if err != nil {
				logger.Error().Err(err).Msg("cannot insert gauge")
				http.Error(w, internalServerError, http.StatusInternalServerError)
				return
			}
			*m.Value, err = a.Storage.SelectGauge(ctx, m.ID)
			if err != nil {
				logger.Error().Err(err).Msg("cannot get gauge")
				return
			}
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				logger.Error().Err(err).Msg("cannot encode gauge")
				return
			}

			w.WriteHeader(http.StatusOK)
			logger.Debug().Msg(sending200OK)

		case models.Counter:
			err := a.Storage.InsertCounter(ctx, m.ID, *m.Delta)
			if err != nil {
				logger.Error().Err(err).Msg("cannot insert counter")
				http.Error(w, internalServerError, http.StatusInternalServerError)
				return
			}

			*m.Delta, err = a.Storage.SelectCounter(ctx, m.ID)
			if err != nil {
				logger.Error().Err(err).Msg("cannot get counter")
				return
			}
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err = enc.Encode(m); err != nil {
				logger.Error().Err(err).Msg("cannot encode counter")
				return
			}

			w.WriteHeader(http.StatusOK)
			logger.Debug().Msg(sending200OK)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func GetTheMetricWithJSON(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "GetTheMetricWithJSON").Logger()
		if r.Header.Get(contentType) != applicationJSON {
			http.Error(w, invalitContentTypeNotJSON, http.StatusBadRequest)
			logger.Debug().Msg(invalidContentType)
			return
		}
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			logger.Error().Err(err).Msg("cannot decode metric")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch m.MType {
		case models.Gauge:
			value, err := a.Storage.SelectGauge(ctx, m.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set(contentType, applicationJSON)
				return
			}
			m.Value = &value
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				logger.Error().Err(err).Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			logger.Debug().Msg(sending200OK)

		case models.Counter:
			delta, err := a.Storage.SelectCounter(ctx, m.ID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set(contentType, applicationJSON)
				return
			}
			m.Delta = &delta
			w.Header().Set(contentType, applicationJSON)
			enc := json.NewEncoder(w)
			if err := enc.Encode(m); err != nil {
				logger.Error().Err(err).Msg("")
				return
			}

			w.WriteHeader(http.StatusOK)
			logger.Debug().Msg("GetTheMetricWithJSON: sending HTTP 200 response")
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func UpdateSliceOfMetrics(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "UpdateSliceOfMetrics").Logger()
		if r.Header.Get(contentType) != applicationJSON {
			http.Error(w, invalitContentTypeNotJSON, http.StatusBadRequest)
			logger.Debug().Msg(invalidContentType)
			return
		}
		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Error().Err(err).Msg("cannot decode slice of metrics")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, m := range metrics {
			if !(m.MType == models.Gauge || m.MType == models.Counter) {
				fmt.Println(m.MType)
				http.Error(w, "Invalid metric type", http.StatusBadRequest)
				return
			}
		}
		if err := a.Storage.InsertBatch(ctx, metrics); err != nil {
			logger.Error().Err(err).Msg("cannot insert batch in handler")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		logger.Debug().Msg(sending200OK)
	}
}

func PingDB(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.Log.With().Str("func", "PingDB").Logger()
		if err := a.Storage.Ping(ctx); err != nil {
			logger.Error().Err(err)
			http.Error(w, internalServerError, http.StatusInternalServerError)
		}
	}
}
