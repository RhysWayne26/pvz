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

var _ services.OrderService = (*TracingOrderService)(nil)

// TracingOrderService is a decorator for OrderService that instruments operations with tracing capabilities. It wraps an existing OrderService implementation and adds tracing spans for each method invocation.
type TracingOrderService struct {
	inner  services.OrderService
	tracer trace.Tracer
}

// NewTracingOrderService creates a new instance of TracingOrderService, wrapping an existing OrderService with tracing capabilities.
func NewTracingOrderService(inner services.OrderService, tracer trace.Tracer) *TracingOrderService {
	return &TracingOrderService{
		inner:  inner,
		tracer: tracer,
	}
}

// AcceptOrder processes an order acceptance request and returns the created or updated order details and any potential error.
func (t TracingOrderService) AcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (models.Order, error) {
	ctx, span := t.tracer.Start(ctx, "OrderService.AcceptOrder",
		trace.WithAttributes(
			attribute.String("order.order_id", strconv.FormatUint(req.OrderID, 10)),
			attribute.String("order.user_id", strconv.FormatUint(req.UserID, 10)),
			attribute.String("order.package", req.Package.String()),
			attribute.Float64("order.weight", float64(req.Weight)),
			attribute.Float64("order.price", float64(req.Price)),
		),
	)
	defer span.End()
	order, err := t.inner.AcceptOrder(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return order, err
}

// IssueOrders submits a batch of orders for issuance and returns their processing results or an error if any issues occur.
func (t TracingOrderService) IssueOrders(ctx context.Context, req requests.IssueOrdersRequest) ([]models.BatchEntryProcessedResult, error) {
	ctx, span := t.tracer.Start(ctx, "OrderService.IssueOrders",
		trace.WithAttributes(
			attribute.Int("orders.count", len(req.OrderIDs)),
		),
	)
	defer span.End()
	results, err := t.inner.IssueOrders(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return results, err
}

// ListOrders retrieves a filtered list of orders, along with pagination details and potential error information.
func (t TracingOrderService) ListOrders(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, uint64, int, error) {
	var attrs []attribute.KeyValue
	if filter.UserID != nil {
		attrs = append(attrs, attribute.Int64("filter.user_id", int64(*filter.UserID)))
	}
	if filter.InPvz != nil {
		attrs = append(attrs, attribute.Bool("filter.in_pvz", *filter.InPvz))
	}
	ctx, span := t.tracer.Start(ctx, "OrderService.ListOrders", trace.WithAttributes(attrs...))
	defer span.End()
	orders, nextID, total, err := t.inner.ListOrders(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return orders, nextID, total, err
}

// CreateClientReturns processes a client returns request and returns a list of batch entry results along with any error.
func (t TracingOrderService) CreateClientReturns(ctx context.Context, req requests.ClientReturnsRequest) ([]models.BatchEntryProcessedResult, error) {
	ctx, span := t.tracer.Start(ctx, "OrderService.CreateClientReturns",
		trace.WithAttributes(
			attribute.Int("orders.count", len(req.OrderIDs)),
		),
	)
	defer span.End()
	results, err := t.inner.CreateClientReturns(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return results, err
}

// ReturnToCourier processes a request to return an order to the courier and records tracing details for the operation.
func (t TracingOrderService) ReturnToCourier(ctx context.Context, req requests.ReturnOrderRequest) error {
	ctx, span := t.tracer.Start(ctx, "OrderService.ReturnToCourier",
		trace.WithAttributes(
			attribute.String("order.order_id", strconv.FormatUint(req.OrderID, 10)),
		),
	)
	defer span.End()
	err := t.inner.ReturnToCourier(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// ListReturns retrieves a list of returned orders matching the specified filter and records tracing for the operation.
func (t TracingOrderService) ListReturns(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, error) {
	var attrs []attribute.KeyValue
	if filter.Status != nil {
		attrs = append(attrs, attribute.String("filter.status", string(*filter.Status)))
	}
	ctx, span := t.tracer.Start(ctx, "OrderService.ListReturns", trace.WithAttributes(attrs...))
	defer span.End()
	orders, _, _, err := t.inner.ListOrders(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return orders, nil
}

// ImportOrders processes a batch import of orders and returns the results of the operation or an error if one occurs.
func (t TracingOrderService) ImportOrders(ctx context.Context, req requests.ImportOrdersRequest) ([]models.BatchEntryProcessedResult, error) {
	ctx, span := t.tracer.Start(ctx, "OrderService.ImportOrders",
		trace.WithAttributes(attribute.Int("orders.count", len(req.Statuses))),
	)
	defer span.End()
	results, err := t.inner.ImportOrders(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return results, err
}
