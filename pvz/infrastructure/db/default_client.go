package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"time"
)

// Mode represents the operational mode, such as read or write, for database queries or commands.
type Mode string

const (
	// ReadMode specifies the operational mode for read-only database queries or commands.
	ReadMode Mode = "read"
	// WriteMode specifies the operational mode for write-only database queries or commands.
	WriteMode Mode = "write"
)

var _ Client = (*DefaultClient)(nil)

// DefaultClient provides a client for managing read and write database connections.
type DefaultClient struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

// NewClient opens 2 connections: for read and write purposes
func NewClient(readDSN, writeDSN string) (*DefaultClient, error) {
	writeDB, err := sql.Open("postgres", writeDSN)
	if err != nil {
		return nil, err
	}
	if err := writeDB.Ping(); err != nil {
		return nil, err
	}

	readDB, err := sql.Open("postgres", readDSN)
	if err != nil {
		return nil, err
	}
	if err := readDB.Ping(); err != nil {
		return nil, err
	}

	return &DefaultClient{
		readDB:  readDB,
		writeDB: writeDB,
	}, nil
}

func (c *DefaultClient) connection(mode Mode) *sql.DB {
	if mode == ReadMode {
		return c.readDB
	}
	return c.writeDB
}

// ExecCtx executes a SQL query using the provided context, mode, and arguments, returning the result or an error.
func (c *DefaultClient) ExecCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := c.connection(mode).ExecContext(ctx, query, args...)
	slog.InfoContext(ctx, "db.exec",
		"mode", mode,
		"query", query,
		"args", args,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err)

	return result, err
}

// QueryCtx executes a SQL query with the specified mode and arguments, using the provided context, and returns rows or an error.
func (c *DefaultClient) QueryCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (*sql.Rows, error) {
	slog.InfoContext(ctx, "db.query",
		"mode", mode,
		"query", query,
		"args", args,
	)
	return c.connection(mode).QueryContext(ctx, query, args...)
}

// QueryRowCtx executes a query expected to return a single row, using the specified context, mode, and arguments.
func (c *DefaultClient) QueryRowCtx(ctx context.Context, mode Mode, query string, args ...interface{}) *sql.Row {
	slog.InfoContext(ctx, "db.queryRow",
		"mode", mode,
		"query", query,
		"args", args,
	)
	return c.connection(mode).QueryRowContext(ctx, query, args...)
}

// WithTx executes a function within a transaction context, ensuring commit or rollback based on success or error.
func (c *DefaultClient) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := c.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				slog.ErrorContext(ctx, "rollback failed", "original_error", err, "rollback_error", rb)
			}
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(tx)
	return
}

// SetConnectionSettings configures database connection pool settings for both read and write connections.
func (c *DefaultClient) SetConnectionSettings(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) {
	for _, db := range []*sql.DB{c.writeDB, c.readDB} {
		db.SetMaxOpenConns(maxOpen)
		db.SetMaxIdleConns(maxIdle)
		db.SetConnMaxLifetime(maxLifetime)
		db.SetConnMaxIdleTime(maxIdleTime)
	}
}

// Ping checks the connectivity of both write and read database connections using the provided context.
func (c *DefaultClient) Ping(ctx context.Context) error {
	if err := c.writeDB.PingContext(ctx); err != nil {
		return err
	}
	return c.readDB.PingContext(ctx)
}

// Close terminates the connections to both the write and read databases, returning an error if any operation fails.
func (c *DefaultClient) Close() error {
	if err := c.writeDB.Close(); err != nil {
		return err
	}
	return c.readDB.Close()
}
