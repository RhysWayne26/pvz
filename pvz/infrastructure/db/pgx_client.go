package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Mode represents a specific operational mode, such as "read" or "write", typically used for database operations.
type Mode string

const (
	// ReadMode is a constant that represents the "read" operational mode, typically used for read-only database operations.
	ReadMode Mode = "read"
	// WriteMode is a constant that represents the "write" operational mode, typically used for modifying database operations.
	WriteMode Mode = "write"
)

// PGXClient defines an interface for interacting with PostgreSQL using context-aware operations and transaction support.
// Provides methods for executing queries, managing transactions, setting connection settings, and handling connections.
type PGXClient interface {
	ExecCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (pgconn.CommandTag, error)
	QueryCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (pgx.Rows, error)
	QueryRowCtx(ctx context.Context, mode Mode, query string, args ...interface{}) pgx.Row
	WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error
	SetConnectionSettings(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration)
	Ping(ctx context.Context) error
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Close() error
}
