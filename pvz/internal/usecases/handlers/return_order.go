package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleReturnOrder processes return-order command to return order to courier
func (f *DefaultFacadeHandler) HandleReturnOrder(ctx context.Context, req requests.ReturnOrderRequest) (responses.ReturnOrderResponse, error) {
	if ctx.Err() != nil {
		return responses.ReturnOrderResponse{}, ctx.Err()
	}

	if err := f.orderService.ReturnToCourier(ctx, req); err != nil {
		return responses.ReturnOrderResponse{}, err
	}

	f.responsesCache.InvalidatePattern("^ListOrders:")
	f.responsesCache.Invalidate(fmt.Sprintf("OrderHistory:%d", req.OrderID))
	f.metrics.IncOrdersServed(1)
	return responses.ReturnOrderResponse{OrderID: req.OrderID}, nil
}
