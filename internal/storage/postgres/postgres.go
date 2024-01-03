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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ospiem/mcollector/internal/models"
	"github.com/rs/zerolog/log"
)

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
				continue
			}
			return fmt.Errorf("failed to store gauge: %w", err)
		}
		rowsAffectedCount := tag.RowsAffected()
		if rowsAffectedCount != 1 {
			return fmt.Errorf("insertGauge expected one row to be affected, actually affected %d", rowsAffectedCount)
		}
		break
	}

	return nil
}

func (db DB) InsertCounter(ctx context.Context, k string, v int64) error {
	sleepTime := 1 * time.Second
	attempt := 0

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
				continue
			}
			return fmt.Errorf("failed to store counter: %w", err)
		}
		rowsAffectedCount := tag.RowsAffected()
		if rowsAffectedCount != 1 {
			return fmt.Errorf("insertCounter expected one row to be affected, actually affected %d", rowsAffectedCount)
		}
		break
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

func (db DB) GetCounters(ctx context.Context) (map[string]int64, error) {
	rows, err := db.pool.Query(ctx, "SELECT id, counter FROM counters")
	if err != nil {
		return nil, fmt.Errorf("postgres failed to select counters: %w", err)
	}
	defer rows.Close()

	counters := make(map[string]int64)

	for rows.Next() {
		var id string
		var counter int64
		if err := rows.Scan(&id, &counter); err != nil {
			return nil, fmt.Errorf("postgres failed to select counter: %w", err)
		}
		counters[id] = counter
	}

	return counters, nil
}

func (db DB) GetGauges(ctx context.Context) (map[string]float64, error) {
	rows, err := db.pool.Query(ctx, "SELECT id, gauge FROM gauges")
	if err != nil {
		return nil, fmt.Errorf("postgres failed to select gauges: %w", err)
	}
	defer rows.Close()

	gauges := make(map[string]float64)

	for rows.Next() {
		var id string
		var gauge float64
		if err := rows.Scan(&id, &gauge); err != nil {
			return nil, fmt.Errorf("postgres failed to select gauge: %w", err)
		}
		gauges[id] = gauge
	}

	return gauges, nil
}

func (db DB) InsertBatch(ctx context.Context, metrics []models.Metrics) error {
	sleepTime := 1 * time.Second
	attempt := 0

	for {
		tx, err := db.pool.Begin(ctx)
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				log.Error().Err(err).Str("func", "InsertBatch").Msg("cannot rollback tx")
			}
		}()
		if err != nil {
			if !isConnExp(err) {
				return fmt.Errorf("failed to open transaction: %w", err)
			}
			if attempt < retryAttempts {
				log.Error().Err(err).Msgf("%s %v", connPGError, sleepTime)
				time.Sleep(sleepTime)
				sleepTime += repeatFactor * time.Second
				attempt++
				continue
			}
			break
		}

		b := createBatch(metrics)
		batchResults := tx.SendBatch(ctx, b)
		_, err = batchResults.Exec()
		if err != nil {
			return fmt.Errorf("cannot exec batch: %w", err)
		}

		if err := batchResults.Close(); err != nil {
			return fmt.Errorf("insert batch cannot close batchResult: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("cannot commit transaction: %w", err)
		}

		break
	}
	return nil
}

func (db DB) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("cannot ping db: %w", err)
	}
	return nil
}

func createBatch(metrics []models.Metrics) *pgx.Batch {
	b := &pgx.Batch{}
	for _, m := range metrics {
		if m.MType == "counter" {
			sqlStatement := `INSERT INTO counters (id, counter) VALUES ($1, $2)
            		 ON CONFLICT (id) DO UPDATE SET counter = counters.counter + EXCLUDED.counter`

			b.Queue(sqlStatement, m.ID, *m.Delta)
		}

		if m.MType == "gauge" {
			sqlStatement := `INSERT INTO gauges (id, gauge) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET gauge = EXCLUDED.gauge`

			b.Queue(sqlStatement, m.ID, *m.Value)
		}
	}
	return b
}
