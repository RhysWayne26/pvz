package decorators

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strconv"
)

var _ services.HistoryService = (*TracingHistoryService)(nil)

// TracingHistoryService is a wrapper around HistoryService that adds tracing spans for method calls.
type TracingHistoryService struct {
	inner  services.HistoryService
	tracer trace.Tracer
}

// NewTracingHistoryService wraps an existing HistoryService with tracing capabilities using the provided tracer.
func NewTracingHistoryService(inner services.HistoryService, tracer trace.Tracer) *TracingHistoryService {
	return &TracingHistoryService{
		inner:  inner,
		tracer: tracer,
	}
}

// Record adds a history entry for an order and tracks the operation with tracing and error reporting.
func (t *TracingHistoryService) Record(ctx context.Context, e models.HistoryEntry) error {
	ctx, span := t.tracer.Start(ctx, "HistoryService.Record",
		trace.WithAttributes(
			attribute.String("history_entry.event", e.Event.String()),
			attribute.String("history_entry.order_id", strconv.FormatUint(e.OrderID, 10)),
		),
	)
	defer span.End()
	err := t.inner.Record(ctx, e)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// List retrieves a list of history entries for a given order, applying the provided filter criteria.
func (t *TracingHistoryService) List(ctx context.Context, filter requests.OrderHistoryFilter) ([]models.HistoryEntry, error) {
	var attrs []attribute.KeyValue
	if filter.OrderID != nil {
		attrs = append(attrs, attribute.Int64("filter.order_id", int64(*filter.OrderID)))
	}
	ctx, span := t.tracer.Start(ctx, "HistoryService.List", trace.WithAttributes(attrs...))
	defer span.End()
	entries, err := t.inner.List(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return entries, err
}
