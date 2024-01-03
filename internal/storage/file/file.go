package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	memorystorage "github.com/ospiem/mcollector/internal/storage/memory"
	"github.com/rs/zerolog/log"

	"github.com/ospiem/mcollector/internal/models"
)

const filePermission = 0644

type FileStorage struct {
	m               memorystorage.MemStorage
	FileStoragePath string
	Restore         bool
	StoreInterval   time.Duration
}

func New(ctx context.Context, fileStoragePath string,
	restore bool,
	storeInterval time.Duration) (*FileStorage, error) {
	ms := memorystorage.New()

	f := FileStorage{
		m:               *ms,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
		StoreInterval:   storeInterval,
	}

	if f.Restore {
		log.Debug().Msg("append to restore metrics")

		err := f.restoreMetrics(ctx)
		if err != nil {
			log.Error().Err(err).Msg("cannot restore the data")
		}

		log.Debug().Msg("restored metrics")
	}

	if f.StoreInterval > 0 {
		go func() {
			t := time.NewTicker(f.StoreInterval)
			defer t.Stop()

			for range t.C {
				log.Debug().Msg("attempt to flush metrics by ticker")
				err := f.flushMetrics(ctx)
				if err != nil {
					log.Error().Err(err).Msg("cannot flush metrics in time")
				}
			}
		}()
	}
	log.Debug().Msgf("initialize file with %s filepath and %s store interval", f.FileStoragePath, f.StoreInterval)
	return &f, nil
}

func (f *FileStorage) InsertGauge(ctx context.Context, k string, v float64) error {
	if err := f.m.InsertGauge(ctx, k, v); err != nil {
		return fmt.Errorf("InsertGauge: %w", err)
	}
	if f.StoreInterval == 0 {
		log.Debug().Msg("attempt to flush metrics in handler")
		err := f.flushMetrics(ctx)
		if err != nil {
			return fmt.Errorf("cannot flush metrics in handler: %w", err)
		}
	}
	return nil
}

func (f *FileStorage) InsertCounter(ctx context.Context, k string, v int64) error {
	if err := f.m.InsertCounter(ctx, k, v); err != nil {
		return fmt.Errorf("InsertCounter: %w", err)
	}
	if f.StoreInterval == 0 {
		log.Debug().Msg("attempt to flush metrics in handler")
		err := f.flushMetrics(ctx)
		if err != nil {
			return fmt.Errorf("cannot flush metrics in handler: %w", err)
		}
	}
	return nil
}

func (f *FileStorage) SelectGauge(ctx context.Context, k string) (float64, error) {
	v, err := f.m.SelectGauge(ctx, k)
	if err != nil {
		return 0, fmt.Errorf("filestorage: %w", err)
	}
	return v, nil
}

func (f *FileStorage) SelectCounter(ctx context.Context, k string) (int64, error) {
	v, err := f.m.SelectCounter(ctx, k)
	if err != nil {
		return 0, fmt.Errorf("filestorage: %w", err)
	}
	return v, nil
}

func (f *FileStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	c, err := f.m.GetCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("filestorage get counters: %w", err)
	}
	return c, nil
}

func (f *FileStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	c, err := f.m.GetGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("filestorage get gauges: %w", err)
	}
	return c, nil
}

func (f *FileStorage) InsertBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, m := range metrics {
		if m.MType == "counter" {
			if err := f.InsertCounter(ctx, m.ID, *m.Delta); err != nil {
				return fmt.Errorf("cannot save counter %s to the file: %w", m.ID, err)
			}
		}
		if m.MType == "gauge" {
			if err := f.InsertGauge(ctx, m.ID, *m.Value); err != nil {
				return fmt.Errorf("cannot save gauge %s to the file: %w", m.ID, err)
			}
		}
	}
	return nil
}

func (f *FileStorage) Ping(ctx context.Context) error {
	return nil
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, filePermission)
	if err != nil {
		return nil, fmt.Errorf("newProduce: %w", err)
	}

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}
func (p *producer) close() error {
	if err := p.file.Close(); err != nil {
		return fmt.Errorf("producer close: %w", err)
	}
	return nil
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, filePermission)
	if err != nil {
		return nil, fmt.Errorf("newConsumer: %w", err)
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) readMetric() (models.Metrics, error) {
	var metric models.Metrics

	if err := c.decoder.Decode(&metric); err != nil {
		return models.Metrics{}, fmt.Errorf("readMetric: %w", err)
	}
	return metric, nil
}

func (c *consumer) close() error {
	if err := c.file.Close(); err != nil {
		return fmt.Errorf("consumer close: %w", err)
	}
	return nil
}

func (p *producer) writeMetric(metric models.Metrics) error {
	if err := p.encoder.Encode(metric); err != nil {
		return fmt.Errorf("writeMetric: %w", err)
	}
	return nil
}

func (f *FileStorage) flushMetrics(ctx context.Context) error {
	const wrapError = "flush metrics error"

	p, err := newProducer(f.FileStoragePath)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}

	defer func() {
		if err := p.close(); err != nil {
			log.Error().Err(err).Msg("cannot close gzip in compress response")
		}
	}()

	counters, err := f.m.GetCounters(ctx)
	if err != nil {
		return fmt.Errorf("filestorage flusmetrics: %w", err)
	}
	log.Debug().Msg("try to flush counters")
	if err = flushCounters(p, counters); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}

	gauges, err := f.m.GetGauges(ctx)
	if err != nil {
		return fmt.Errorf("filestorage flusmetrics: %w", err)
	}
	log.Debug().Msg("try to flush gauges")
	if err = flushGauges(p, gauges); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	return nil
}

func (f *FileStorage) restoreMetrics(ctx context.Context) error {
	const wrapError = "restore metrics error"

	c, err := newConsumer(f.FileStoragePath)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}

	defer func() {
		if err := c.close(); err != nil {
			log.Error().Err(err).Msg("cannot close gzip in compress response")
		}
	}()

	for {
		metric, err := c.readMetric()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%s: %w", wrapError, err)
		}

		switch metric.MType {
		case models.Counter:
			if err := f.m.InsertCounter(ctx, metric.ID, *metric.Delta); err != nil {
				log.Error().Err(err).Msgf("cannot restore counter %s", metric.ID)
			}
		case models.Gauge:
			if err := f.m.InsertGauge(ctx, metric.ID, *metric.Value); err != nil {
				log.Error().Err(err).Msgf("cannot restore gauge %s", metric.ID)
			}
		}
	}
	return nil
}

func flushCounters(p *producer, c map[string]int64) error {
	const wrapError = "flush counters error"
	m := models.Metrics{MType: models.Counter}
	for i, v := range c {
		m.ID = i
		m.Delta = &v
		if err := p.writeMetric(m); err != nil {
			return fmt.Errorf("%s: %w", wrapError, err)
		}
	}
	return nil
}

func flushGauges(p *producer, c map[string]float64) error {
	const wrapError = "flush counters error"
	m := models.Metrics{MType: models.Gauge}
	for i, v := range c {
		m.ID = i
		m.Value = &v
		if err := p.writeMetric(m); err != nil {
			return fmt.Errorf("%s: %w", wrapError, err)
		}
	}
	return nil
}
