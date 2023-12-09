package storage

import (
	"time"

	"github.com/ilnsm/mcollector/internal/storage/filestorage"
	memorystorage "github.com/ilnsm/mcollector/internal/storage/memory"
)

type Storage interface {
	InsertGauge(k string, v float64) error
	InsertCounter(k string, v int64) error
	SelectGauge(k string) (float64, error)
	SelectCounter(k string) (int64, error)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
}

func New(fileStoragePath string,
	restore bool,
	storeInterval time.Duration) (Storage, error) {
	if fileStoragePath != "" {
		f, err := filestorage.New(fileStoragePath, restore, storeInterval)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	m, err := memorystorage.New()
	if err != nil {
		return nil, err
	}
	return m, nil
}
