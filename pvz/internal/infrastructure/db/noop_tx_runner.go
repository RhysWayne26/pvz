package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type NoOpTxRunner struct{}

func NewNoOpTxRunner() *NoOpTxRunner {
	return &NoOpTxRunner{}
}

func (r *NoOpTxRunner) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}
