package memorystorage

import (
	"context"
	"errors"
)

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func New() *MemStorage {
	s := MemStorage{make(map[string]int64), make(map[string]float64)}
	return &s
}

func (m *MemStorage) InsertGauge(ctx context.Context, k string, v float64) error {
	m.gauge[k] = v
	return nil
}
func (m *MemStorage) InsertCounter(ctx context.Context, k string, v int64) error {
	m.counter[k] += v
	return nil
}

func (m *MemStorage) SelectGauge(ctx context.Context, k string) (float64, error) {
	if v, ok := m.gauge[k]; ok {
		return v, nil
	}
	return 0, errors.New("gauge does not exist")
}

func (m *MemStorage) SelectCounter(ctx context.Context, k string) (int64, error) {
	if v, ok := m.counter[k]; ok {
		return v, nil
	}
	return 0, errors.New("counter does not exist")
}

func (m *MemStorage) GetCounters(ctx context.Context) map[string]int64 {
	return m.counter
}
func (m *MemStorage) GetGauges(ctx context.Context) map[string]float64 {
	return m.gauge
}

func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}
