package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	return &DB{
		pool: pool,
	}, nil
}

func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pgConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
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

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (d DB) InsertGauge(ctx context.Context, k string, v float64) error {
	// TODO implement me
	panic("")
}

func (d DB) InsertCounter(ctx context.Context, k string, v int64) error {
	// TODO implement me
	panic("")
}

func (d DB) SelectGauge(ctx context.Context, k string) (float64, error) {
	// TODO implement me
	panic("")
}

func (d DB) SelectCounter(ctx context.Context, k string) (int64, error) {
	// TODO implement me
	panic("")
}

func (d DB) GetCounters(ctx context.Context) map[string]int64 {
	// TODO implement me
	panic("")
}

func (d DB) GetGauges(ctx context.Context) map[string]float64 {
	// TODO implement me
	panic("")
}

func (d DB) Ping(ctx context.Context) error {
	if err := d.pool.Ping(ctx); err != nil {
		return fmt.Errorf("cannot ping db: %w", err)
	}
	return nil
}
