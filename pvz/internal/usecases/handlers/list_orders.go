package handlers

import (
	"context"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleListOrders processes the list-orders request and returns the result.
func (f *DefaultFacadeHandler) HandleListOrders(ctx context.Context, req requests.OrdersFilterRequest) (responses.ListOrdersResponse, error) {
	if ctx.Err() != nil {
		return responses.ListOrdersResponse{}, ctx.Err()
	}

	orders, nextID, total, err := f.OrderService.ListOrders(ctx, req)
	if err != nil {
		return responses.ListOrdersResponse{}, err
	}

	return responses.ListOrdersResponse{
		Orders: orders,
		NextID: &nextID,
		Total:  utils.Ptr(total),
	}, nil
}
