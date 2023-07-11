package postgres

import (
	"context"
	"fmt"
	"github.com/OkDenAl/mbtiles_converter/pkg/logging"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const (
	maxDefaultAttempts = 10
	defaultSleepConn   = time.Second
)

type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func New(dsn string, maxPoolConns int, log logging.Logger) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	attemptsCount := maxDefaultAttempts
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error while parse pgpool config: %w", err)
	}
	cfg.MaxConns = int32(maxPoolConns)
	for attemptsCount > 0 {
		pool, err = pgxpool.NewWithConfig(context.Background(), cfg)
		if err == nil {
			break
		}
		log.Infof("try to connect to postgres... attempts left %d\n", attemptsCount)
		time.Sleep(defaultSleepConn)
		attemptsCount--
	}
	if err != nil {
		return nil, fmt.Errorf("error while connect to postgres: %w", err)
	}
	return pool, nil
}
