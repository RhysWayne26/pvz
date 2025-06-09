package handlers

import (
	"context"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleListOrders processes the list-orders request and returns the result.
func (f *DefaultFacadeHandler) HandleListOrders(ctx context.Context, req requests.OrdersFilterRequest) (responses.ListOrdersResponse, error) {
	select {
	case <-ctx.Done():
		return responses.ListOrdersResponse{}, ctx.Err()
	default:
	}

	orders, nextID, total, err := f.orderService.ListOrders(req)
	if err != nil {
		return responses.ListOrdersResponse{}, err
	}

	return responses.ListOrdersResponse{
		Orders: orders,
		NextID: &nextID,
		Total:  utils.Ptr(total),
	}, nil
}
