package memorystorage

import "errors"

type MemStorage struct {
	counter map[string]int64
	gauge   map[string]float64
}

func New() (*MemStorage, error) {
	s := MemStorage{make(map[string]int64), make(map[string]float64)}
	return &s, nil
}

func (m *MemStorage) InsertGauge(k string, v float64) error {
	m.gauge[k] = v
	return nil
}
func (m *MemStorage) InsertCounter(k string, v int64) error {
	m.counter[k] += v
	return nil
}

func (m *MemStorage) SelectGauge(k string) (float64, error) {
	if v, ok := m.gauge[k]; ok {
		return v, nil
	}
	return 0, errors.New("gauge does not exist")
}

func (m *MemStorage) SelectCounter(k string) (int64, error) {
	if v, ok := m.counter[k]; ok {
		return v, nil
	}
	return 0, errors.New("counter does not exist")
}

func (m *MemStorage) GetCounters() map[string]int64 {
	return m.counter
}
func (m *MemStorage) GetGauges() map[string]float64 {
	return m.gauge
}
