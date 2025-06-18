package db

import (
	"context"
	"database/sql"
	"time"
)

// Client defines an interface for executing SQL queries, managing transactions, and configuring database connections.
type Client interface {
	ExecCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (sql.Result, error)
	QueryCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowCtx(ctx context.Context, mode Mode, query string, args ...interface{}) *sql.Row
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	SetConnectionSettings(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration)
	Ping(ctx context.Context) error
	Close() error
}
