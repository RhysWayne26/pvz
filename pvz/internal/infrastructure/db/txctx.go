package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type txKeyType struct{}

var txKey = txKeyType{}

// WithTxContext returns a new context with the pgx.Tx transaction associated using a predefined key.
func WithTxContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext retrieves a pgx.Tx object and a boolean from the context if one is stored under the predefined key.
func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
