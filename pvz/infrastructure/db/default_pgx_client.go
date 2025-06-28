package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

var _ PGXClient = (*DefaultPGXClient)(nil)

// DefaultPGXClient provides a default implementation of the PGXClient interface for interacting with a PostgreSQL database.
// It manages separate connection pools for read and write operations.
type DefaultPGXClient struct {
	ReadPool  *pgxpool.Pool
	WritePool *pgxpool.Pool
}

// NewDefaultPGXClient initializes a DefaultPGXClient with separate read and write database connection pools.
func NewDefaultPGXClient(readDSN, writeDSN string) (*DefaultPGXClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	readPool, err := pgxpool.New(ctx, readDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to read DB: %w", err)
	}
	if err := readPool.Ping(ctx); err != nil {
		readPool.Close()
		return nil, fmt.Errorf("read DB ping failed: %w", err)
	}
	writePool, err := pgxpool.New(ctx, writeDSN)
	if err != nil {
		readPool.Close()
		return nil, fmt.Errorf("failed to connect to write DB: %w", err)
	}
	if err := writePool.Ping(ctx); err != nil {
		readPool.Close()
		writePool.Close()
		return nil, fmt.Errorf("write DB ping failed: %w", err)
	}
	return &DefaultPGXClient{
		ReadPool:  readPool,
		WritePool: writePool,
	}, nil
}

// ExecCtx executes a query in the specified mode (read or write) using a context and returns the command tag and error.
// It logs query details, execution time, rows affected, and any error encountered during execution.
func (c *DefaultPGXClient) ExecCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	pool := c.selectPool(mode)
	tag, err := pool.Exec(ctx, query, args...)
	slog.InfoContext(ctx, "db.exec",
		"mode", mode,
		"query", query,
		"duration_ms", time.Since(start).Milliseconds(),
		"rows_affected", tag.RowsAffected(),
		"error", err,
	)

	return tag, err
}

// QueryCtx executes a database query in the specified mode using a context and returns the result rows and any error.
func (c *DefaultPGXClient) QueryCtx(ctx context.Context, mode Mode, query string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	pool := c.selectPool(mode)
	rows, err := pool.Query(ctx, query, args...)
	slog.InfoContext(ctx, "db.query",
		"mode", mode,
		"query", query,
		"duration_ms", time.Since(start).Milliseconds(),
		"error", err,
	)
	return rows, err
}

// QueryRowCtx executes a query in the specified mode using a context and returns a single result row.
func (c *DefaultPGXClient) QueryRowCtx(ctx context.Context, mode Mode, query string, args ...interface{}) pgx.Row {
	start := time.Now()
	pool := c.selectPool(mode)
	row := pool.QueryRow(ctx, query, args...)
	slog.InfoContext(ctx, "db.queryRow",
		"mode", mode,
		"query", query,
		"duration_ms", time.Since(start).Milliseconds())
	return row
}

// WithTx executes a function within a database transaction, ensuring proper commit or rollback handling.
func (c *DefaultPGXClient) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := c.WritePool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				slog.ErrorContext(ctx, "rollback after panic failed", "panic", p, "rollback_error", rbErr)
			}
			panic(p)
		}
	}()
	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			slog.ErrorContext(ctx, "rollback failed", "original_error", err, "rollback_error", rbErr)
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Ping checks the connectivity of both read and write database connection pools, returning an error if any ping fails.
func (c *DefaultPGXClient) Ping(ctx context.Context) error {
	if err := c.ReadPool.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "db.ping", "mode", ReadMode, "error", err)
		return fmt.Errorf("read pool ping failed: %w", err)
	}
	if err := c.WritePool.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "db.ping", "mode", WriteMode, "error", err)
		return fmt.Errorf("write pool ping failed: %w", err)
	}
	return nil
}

// SetConnectionSettings configures connection pool settings including max open connections, idle connections, and lifetimes.
func (c *DefaultPGXClient) SetConnectionSettings(maxOpen, maxIdle int, maxLifetime, maxIdleTime time.Duration) {
	c.WritePool.Config().MaxConns = int32(maxOpen)
	c.WritePool.Config().MinConns = int32(maxIdle)
	c.WritePool.Config().MaxConnLifetime = maxLifetime
	c.WritePool.Config().MaxConnIdleTime = maxIdleTime
	c.ReadPool.Config().MaxConns = int32(maxOpen)
	c.ReadPool.Config().MinConns = int32(maxIdle)
	c.ReadPool.Config().MaxConnLifetime = maxLifetime
	c.ReadPool.Config().MaxConnIdleTime = maxIdleTime
}

// Query executes a read-only database query using the specified context, query string, and arguments. Returns rows and error.
func (c *DefaultPGXClient) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return c.QueryCtx(ctx, ReadMode, query, args...)
}

// QueryRow executes a query in read mode, returning a single result row using the provided context, query, and arguments.
func (c *DefaultPGXClient) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return c.QueryRowCtx(ctx, ReadMode, query, args...)
}

// Exec executes a write operation in the database using the provided context, query, and arguments. Returns a command tag and error.
func (c *DefaultPGXClient) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.ExecCtx(ctx, WriteMode, query, args...)
}

// Close safely closes the read and write database connection pools, recovering from any panic that may occur.
func (c *DefaultPGXClient) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during pool close: %v", r)
		}
	}()
	if c.ReadPool != nil {
		c.ReadPool.Close()
	}
	if c.WritePool != nil {
		c.WritePool.Close()
	}
	return nil
}

func (c *DefaultPGXClient) selectPool(mode Mode) *pgxpool.Pool {
	if mode == ReadMode {
		return c.ReadPool
	}
	return c.WritePool
}
