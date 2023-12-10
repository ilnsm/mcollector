package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"
	"github.com/rs/zerolog/log"

	"github.com/ilnsm/mcollector/internal/models"
)

const filePermission = 0600

type Storage interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	SelectGauge(k string) (float64, error)
	SelectCounter(k string) (int64, error)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
}

type FileStorage struct {
	m               memorystorage.MemStorage
	FileStoragePath string
	Restore         bool
	StoreInterval   time.Duration
}

func New(fileStoragePath string,
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

		err := f.restoreMetrics()
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
				err := f.flushMetrics()
				if err != nil {
					log.Error().Err(err).Msg("cannot flush metrics in time")
				}
			}
		}()
	}
	log.Debug().Msgf("initialize file with %s filepath and %s store interval", f.FileStoragePath, f.StoreInterval)
	return &f, nil
}

func (f *FileStorage) InsertGauge(k string, v float64) error {
	if err := f.m.InsertGauge(k, v); err != nil {
		return fmt.Errorf("InsertGauge: %w", err)
	}
	if f.StoreInterval == 0 {
		log.Debug().Msg("attempt to flush metrics in handler")
		err := f.flushMetrics()
		if err != nil {
			return fmt.Errorf("cannot flush metrics in handler: %w", err)
		}
	}
	return nil
}

func (f *FileStorage) InsertCounter(k string, v int64) error {
	if err := f.m.InsertCounter(k, v); err != nil {
		return fmt.Errorf("InsertCounter: %w", err)
	}
	if f.StoreInterval == 0 {
		log.Debug().Msg("attempt to flush metrics in handler")
		err := f.flushMetrics()
		if err != nil {
			return fmt.Errorf("cannot flush metrics in handler: %w", err)
		}
	}
	return nil
}

func (f *FileStorage) SelectGauge(k string) (float64, error) {
	v, err := f.m.SelectGauge(k)
	if err != nil {
		return 0, fmt.Errorf("filestorage: %w", err)
	}
	return v, nil
}

func (f *FileStorage) SelectCounter(k string) (int64, error) {
	v, err := f.m.SelectCounter(k)
	if err != nil {
		return 0, fmt.Errorf("filestorage: %w", err)
	}
	return v, nil
}

func (f *FileStorage) GetCounters() map[string]int64 {
	c := f.m.GetCounters()
	return c
}

func (f *FileStorage) GetGauges() map[string]float64 {
	c := f.m.GetGauges()
	return c
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

func (f *FileStorage) flushMetrics() error {
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

	counters := f.m.GetCounters()
	log.Debug().Msg("try to flush counters")
	if err = flushCounters(p, counters); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}

	gauges := f.m.GetGauges()
	log.Debug().Msg("try to flush gauges")
	if err = flushGauges(p, gauges); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	return nil
}

func (f *FileStorage) restoreMetrics() error {
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
			if err := f.m.InsertCounter(metric.ID, *metric.Delta); err != nil {
				log.Error().Err(err).Msgf("cannot restore counter %s", metric.ID)
			}
		case models.Gauge:
			if err := f.m.InsertGauge(metric.ID, *metric.Value); err != nil {
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
