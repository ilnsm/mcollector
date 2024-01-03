package storage

import (
	"context"
	"fmt"

	"github.com/ospiem/mcollector/internal/models"
	"github.com/ospiem/mcollector/internal/server/config"
	"github.com/ospiem/mcollector/internal/storage/file"
	memorystorage "github.com/ospiem/mcollector/internal/storage/memory"
	"github.com/ospiem/mcollector/internal/storage/postgres"
)

type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) (map[string]int64, error)
	GetGauges(ctx context.Context) (map[string]float64, error)
	InsertBatch(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

func New(ctx context.Context, cfg config.Config) (Storage, error) {
	if cfg.DatabaseDsn != "" {
		db, err := postgres.NewDB(ctx, cfg.DatabaseDsn)
		if err != nil {
			return nil, fmt.Errorf("failed to init postgres pool: %w", err)
		}
		return db, nil
	}

	if cfg.FileStoragePath != "" {
		f, err := file.New(ctx, cfg.FileStoragePath, cfg.Restore, cfg.StoreInterval)
		if err != nil {
			return nil, fmt.Errorf("new storage error: %w", err)
		}
		return f, nil
	}

	return memorystorage.New(), nil
}
