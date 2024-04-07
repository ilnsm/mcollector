// Package transport provides functionality for handling HTTP transport layer.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/server/middleware/compress"
	"github.com/ospiem/mcollector/internal/server/middleware/hash"
	"github.com/ospiem/mcollector/internal/server/middleware/logger"
	"github.com/ospiem/mcollector/internal/server/middleware/ssl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// htmlTemplate represents the HTML template for displaying metrics data.
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

// Constants related to HTTP headers and status codes.
const (
	contentType               = "Content-Type"
	applicationJSON           = "application/json"
	internalServerError       = "Internal server error"
	sending200OK              = "sending HTTP 200 response"
	invalidContentType        = "invalid content-type"
	invalitContentTypeNotJSON = "Invalid Content-Type, expected application/json"
)

// Storage is an interface that defines methods for interacting with storage.
// It includes methods for inserting, selecting, and retrieving metrics, as well as for pinging and closing the storage.
type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
	GetGauges(ctx context.Context) (map[string]float64, error)
	InsertBatch(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

// API represents an HTTP API server. It includes a storage interface, a logger, and a server configuration.
type API struct {
	Storage Storage        // Storage is the storage interface implemention.
	Log     zerolog.Logger // Log is the logger instance.
	Cfg     config.Config  // Cfg is the server configuration.
}

// New creates a new instance of the API server. It takes a server configuration, a storage interface, and a logger as parameters.
func New(cfg *config.Config, s Storage, l *zerolog.Logger) *API {
	return &API{
		Cfg:     *cfg,
		Storage: s,
		Log:     *l,
	}
}

// registerAPI registers the API routes and their corresponding handlers. It also sets up the necessary middleware for each route.
func (a *API) registerAPI() chi.Router {
	// Parse the private key from the server configuration.
	privateKey, err := ssl.ParsePrivateKey(a.Cfg.CryptoKey)
	if err != nil {
		a.Log.Fatal().Msg("failed to parse private key")
	}

	// Create a new router.
	r := chi.NewRouter()

	// Set up the middleware for the router.
	r.Use(middleware.Recoverer)
	r.Use(logger.RequestLogger(a.Log))

	// Mount the profiler endpoint for debugging purposes.
	r.Mount("/debug", middleware.Profiler())

	// Define the routes for updating metrics.
	r.Route("/", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// Set up the middleware for updating metrics.
			r.Use(compress.DecompressRequest(a.Log))
			r.Use(hash.VerifyRequestBodyIntegrity(a.Log, a.Cfg.Key))
			r.Use(ssl.Terminate(a.Log, privateKey))
			r.Use(compress.CompressResponse(a.Log))

			// Define the routes for updating a single metric.
			r.Route("/update", func(r chi.Router) {
				r.Post("/", UpdateTheMetricWithJSON(a))
				r.Post("/{mType}/{mName}/{mValue}", UpdateTheMetric(a))
			})

			// Define the route for updating a slice of metrics.
			r.Post("/updates/", UpdateSliceOfMetrics(a))
		})
	})

	// Define the routes for getting metrics.
	r.Group(func(r chi.Router) {
		// Set up the middleware for getting metrics.
		r.Use(compress.DecompressRequest(a.Log))

		// Define the route for listing all metrics.
		r.Get("/", ListAllMetrics(a))

		// Define the routes for getting a single metric value.
		r.Route("/value", func(r chi.Router) {
			r.Post("/", GetTheMetricWithJSON(a))
			r.Get("/{mType}/{mName}", GetTheMetric(a))
		})

		// Define the route for pinging the database.
		r.Get("/ping", PingDB(a))
	})

	return r
}

// InitServer initializes the server with the registered API routes. It returns an HTTP server.
func (a *API) InitServer() *http.Server {
	a.Log.Info().Msgf("Starting server on %s", a.Cfg.Endpoint)

	// Register the API routes.
	r := a.registerAPI()

	// Return a new HTTP server.
	return &http.Server{
		Addr:    a.Cfg.Endpoint,
		Handler: r,
	}
}

// UpdateTheMetric handles updating a metric based on the HTTP request.
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
					http.Error(w, "Not Found", http.StatusBadRequest)
				}

				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

// GetTheMetric retrieves a metric based on the HTTP request.
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

// ListAllMetrics lists all metrics in a human-readable HTML format.
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
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		g, err := a.Storage.GetGauges(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get gauges")
			http.Error(w, "", http.StatusInternalServerError)
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

// UpdateTheMetricWithJSON handles updating a metric using JSON format.
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

// GetTheMetricWithJSON retrieves a metric using JSON format.
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

// UpdateSliceOfMetrics handles updating a slice of metrics using JSON format.
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

// PingDB pings the database to check its connectivity.
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
