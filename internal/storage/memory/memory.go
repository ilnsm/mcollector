package memorystorage

import (
	"context"
	"errors"

	"github.com/ilnsm/mcollector/internal/models"
)

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func New() *MemStorage {
	s := MemStorage{make(map[string]int64), make(map[string]float64)}
	return &s
}

func (mem *MemStorage) InsertGauge(ctx context.Context, k string, v float64) error {
	mem.gauge[k] = v
	return nil
}
func (mem *MemStorage) InsertCounter(ctx context.Context, k string, v int64) error {
	mem.counter[k] += v
	return nil
}

func (mem *MemStorage) SelectGauge(ctx context.Context, k string) (float64, error) {
	if v, ok := mem.gauge[k]; ok {
		return v, nil
	}
	return 0, errors.New("gauge does not exist")
}

func (mem *MemStorage) SelectCounter(ctx context.Context, k string) (int64, error) {
	if v, ok := mem.counter[k]; ok {
		return v, nil
	}
	return 0, errors.New("counter does not exist")
}

func (mem *MemStorage) GetCounters(ctx context.Context) (map[string]int64, error) {
	return mem.counter, nil
}
func (mem *MemStorage) GetGauges(ctx context.Context) (map[string]float64, error) {
	return mem.gauge, nil
}

func (mem *MemStorage) InsertBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, m := range metrics {
		if m.MType == "counter" {
			mem.counter[m.ID] += *m.Delta
		}
		if m.MType == "gauge" {
			mem.gauge[m.ID] = *m.Value
		}
	}
	return nil
}
func (mem *MemStorage) Ping(ctx context.Context) error {
	return nil
}
