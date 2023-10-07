package memorystorage

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
