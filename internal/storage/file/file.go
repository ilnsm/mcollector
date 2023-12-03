package file

import (
	"encoding/json"
	"fmt"
	"github.com/ilnsm/mcollector/internal/models"
	"github.com/rs/zerolog"
	"io"
	"os"
)

type Storage interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	SelectGauge(k string) (float64, error)
	SelectCounter(k string) (int64, error)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil

}
func (p *producer) close() error {
	return p.file.Close()
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) readMetric() (models.Metrics, error) {
	var metric models.Metrics

	if err := c.decoder.Decode(&metric); err != nil {
		return models.Metrics{}, err
	}
	return metric, nil
}

func (c *consumer) close() error {
	return c.file.Close()
}

func (p *producer) writeMetric(metric models.Metrics) error {
	if err := p.encoder.Encode(metric); err != nil {
		return err
	}
	return nil
}

func FlushMetrics(s Storage, filename string) error {
	const wrapError = "flush metrics error"

	p, err := newProducer(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	defer p.close()

	counters := s.GetCounters()
	if err = flushCounters(p, counters); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}

	gauges := s.GetGauges()
	if err = flushGauges(p, gauges); err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	return nil
}

func RestoreMetrics(s Storage, filename string, l zerolog.Logger) error {
	const wrapError = "restore metrics error"

	c, err := newConsumer(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", wrapError, err)
	}
	defer c.close()

	for {
		metric, err := c.readMetric()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("%s: %w", wrapError, err)
		}

		switch metric.MType {
		case models.Counter:
			if err := s.InsertCounter(metric.ID, *metric.Delta); err != nil {
				l.Error().Err(err).Msgf("cannot restore counter %s", metric.ID)
			}
		case models.Gauge:
			if err := s.InsertGauge(metric.ID, *metric.Value); err != nil {
				l.Error().Err(err).Msgf("cannot restore gauge %s", metric.ID)
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
