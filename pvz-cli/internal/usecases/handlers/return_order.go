package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleReturnOrder processes return-order command to return order to courier
func (f *DefaultFacadeHandler) HandleReturnOrder(ctx context.Context, req requests.ReturnOrderRequest) (responses.ReturnOrderResponse, error) {
	select {
	case <-ctx.Done():
		return responses.ReturnOrderResponse{}, ctx.Err()
	default:
	}

	if err := f.orderService.ReturnToCourier(req); err != nil {
		return responses.ReturnOrderResponse{}, err
	}

	return responses.ReturnOrderResponse{OrderID: req.OrderID}, nil
}
