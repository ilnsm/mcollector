package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ilnsm/mcollector/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	InsertGauge(ctx context.Context, k string, v float64) error
	InsertCounter(ctx context.Context, k string, v int64) error
	SelectGauge(ctx context.Context, k string) (float64, error)
	SelectCounter(ctx context.Context, k string) (int64, error)
	GetCounters(ctx context.Context) map[string]int64
	GetGauges(ctx context.Context) map[string]float64
	InsertBatch(ctx context.Context, metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

const connPGError = "cannot connect to postgres, will retry in"
const retryAttempts = 3
const repeatFactor = 2

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

func (db DB) InsertGauge(ctx context.Context, k string, v float64) error {
	sleepTime := 1 * time.Second
	attempt := 0

requestLoop:
	for {
		tag, err := db.pool.Exec(
			ctx,
			`INSERT INTO gauges (id, gauge) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET gauge = EXCLUDED.gauge`,
			k, v,
		)
		if err != nil {
			if attempt < retryAttempts {
				log.Error().Err(err).Msgf("%s %v", connPGError, sleepTime)
				time.Sleep(sleepTime)
				sleepTime += repeatFactor * time.Second
				attempt++
				continue requestLoop
			}
			return fmt.Errorf("failed to store gauge: %w", err)
		}
		rowsAffectedCount := tag.RowsAffected()
		if rowsAffectedCount != 1 {
			return fmt.Errorf("insertGauge expected one row to be affected, actually affected %d", rowsAffectedCount)
		}
		break requestLoop
	}

	return nil
}

func (db DB) InsertCounter(ctx context.Context, k string, v int64) error {
	sleepTime := 1 * time.Second
	attempt := 0

requestLoop:
	for {
		tag, err := db.pool.Exec(
			ctx,
			`INSERT INTO counters (id, counter) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET counter = counters.counter + EXCLUDED.counter`,
			k, v,
		)
		if err != nil {
			if attempt < retryAttempts {
				log.Error().Err(err).Msgf("%s %v", connPGError, sleepTime)
				time.Sleep(sleepTime)
				sleepTime += repeatFactor * time.Second
				attempt++
				continue requestLoop
			}
			return fmt.Errorf("failed to store counter: %w", err)
		}
		rowsAffectedCount := tag.RowsAffected()
		if rowsAffectedCount != 1 {
			return fmt.Errorf("insertCounter expected one row to be affected, actually affected %d", rowsAffectedCount)
		}
		break requestLoop
	}

	return nil
}

func (db DB) SelectGauge(ctx context.Context, k string) (float64, error) {
	var g float64
	row := db.pool.QueryRow(
		ctx,
		`SELECT gauge FROM gauges WHERE id = $1`,
		k,
	)
	if err := row.Scan(&g); err != nil {
		return 0, fmt.Errorf("failed to select gauge: %w", err)
	}
	return g, nil
}

func (db DB) SelectCounter(ctx context.Context, k string) (int64, error) {
	var c int64
	row := db.pool.QueryRow(
		ctx,
		`SELECT counter FROM counters WHERE id = $1`,
		k,
	)
	if err := row.Scan(&c); err != nil {
		return 0, fmt.Errorf("failed to select counter: %w", err)
	}
	return c, nil
}

func (db DB) GetCounters(ctx context.Context) map[string]int64 {
	rows, err := db.pool.Query(ctx, "SELECT id, counter FROM counters")
	if err != nil {
		return nil
	}
	defer rows.Close()

	counters := make(map[string]int64)

	for rows.Next() {
		var id string
		var counter int64
		if err := rows.Scan(&id, &counter); err != nil {
			return nil
		}
		counters[id] = counter
	}

	return counters
}

func (db DB) GetGauges(ctx context.Context) map[string]float64 {
	rows, err := db.pool.Query(ctx, "SELECT id, gauge FROM gauges")
	if err != nil {
		return nil
	}
	defer rows.Close()

	gauges := make(map[string]float64)

	for rows.Next() {
		var id string
		var gauge float64
		if err := rows.Scan(&id, &gauge); err != nil {
			return nil
		}
		gauges[id] = gauge
	}

	return gauges
}

func (db DB) InsertBatch(ctx context.Context, metrics []models.Metrics) error {
	sleepTime := 1 * time.Second
	attempt := 0

requestLoop:
	for {
		begin, err := db.pool.Begin(ctx)
		if err != nil {
			if !isConnExp(err) {
				return fmt.Errorf("failed to open transaction: %w", err)
			}
			if attempt < retryAttempts {
				log.Error().Err(err).Msgf("%s %v", connPGError, sleepTime)
				time.Sleep(sleepTime)
				sleepTime += repeatFactor * time.Second
				attempt++
				continue requestLoop
			}
			break requestLoop
		}

		for _, m := range metrics {
			if m.MType == "counter" {
				tag, err := begin.Exec(ctx,
					`INSERT INTO counters (id, counter) VALUES ($1, $2)
            		 ON CONFLICT (id) DO UPDATE SET counter = counters.counter + EXCLUDED.counter`,
					m.ID, *m.Delta)
				if err != nil {
					return fmt.Errorf("failed to insert counter from batch: %w", err)
				}
				if rowsAffectedCount := tag.RowsAffected(); rowsAffectedCount != 1 {
					log.Error().Msgf("insertBatch expected one row to be affected, actually affected %d", rowsAffectedCount)
				}
			}

			if m.MType == "gauge" {
				tag, err := begin.Exec(ctx,
					`INSERT INTO gauges (id, gauge) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET gauge = EXCLUDED.gauge`,
					m.ID, *m.Value)
				if err != nil {
					return fmt.Errorf("failed to insert gauge from batch: %w", err)
				}
				if rowsAffectedCount := tag.RowsAffected(); rowsAffectedCount != 1 {
					log.Error().Msgf("insertBatch expected one row to be affected, actually affected %d", rowsAffectedCount)
				}
			}
		}
		if err := begin.Commit(ctx); err != nil {
			return fmt.Errorf("cannot commit transaction: %w", err)
		}
		break requestLoop
	}
	return nil
}

func (db DB) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("cannot ping db: %w", err)
	}
	return nil
}
