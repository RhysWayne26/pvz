package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ TxRunner = (*TracingTxRunner)(nil)

// TracingTxRunner is a wrapper around TxRunner that adds tracing capabilities using the provided tracer.
type TracingTxRunner struct {
	inner  TxRunner
	tracer trace.Tracer
}

// NewTracingTxRunner creates a new TracingTxRunner wrapping an inner TxRunner and adding tracing capabilities via the provided tracer.
func NewTracingTxRunner(inner TxRunner, tracer trace.Tracer) *TracingTxRunner {
	return &TracingTxRunner{
		inner:  inner,
		tracer: tracer,
	}
}

// WithTx executes a function within the context of a database transaction and traces its execution with the provided tracer.
func (t TracingTxRunner) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	ctx, span := t.tracer.Start(ctx, "TxRunner.WithTx")
	defer span.End()
	err := t.inner.WithTx(ctx, fn)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
