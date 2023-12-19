package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) map[string]int64
	GetGauges(ctx context.Context) map[string]float64
	Ping(ctx context.Context) error
}

type Config struct {
	DSN string
}

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, DSN string) (*DB, error) {
	pool, err := initPool(ctx, DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	return &DB{
		pool: pool,
	}, nil
}

func initPool(ctx context.Context, DSN string) (*pgxpool.Pool, error) {

	pgConfig, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db does not ping: %w", err)
	}

	return pool, nil
}

func (D DB) InsertGauge(ctx context.Context, k string, v float64) error {
	//TODO implement me
	panic("implement me")
}

func (D DB) InsertCounter(ctx context.Context, k string, v int64) error {
	//TODO implement me
	panic("implement me")
}

func (D DB) SelectGauge(ctx context.Context, k string) (float64, error) {
	//TODO implement me
	panic("implement me")
}

func (D DB) SelectCounter(ctx context.Context, k string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (D DB) GetCounters(ctx context.Context) map[string]int64 {
	//TODO implement me
	panic("implement me")
}

func (D DB) GetGauges(ctx context.Context) map[string]float64 {
	//TODO implement me
	panic("implement me")
}

func (D DB) Ping(ctx context.Context) error {
	return D.pool.Ping(ctx)
}
