package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleAcceptOrder processes accept-order command with package pricing validation, optionally suppressing output for batch import
func (f *DefaultFacadeHandler) HandleAcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (responses.AcceptOrderResponse, error) {
	if ctx.Err() != nil {
		return responses.AcceptOrderResponse{}, ctx.Err()
	}

	order, err := f.OrderService.AcceptOrder(ctx, req)
	if err != nil {
		return responses.AcceptOrderResponse{}, err
	}

	return responses.AcceptOrderResponse{
		OrderID: order.OrderID,
		Package: order.Package,
		Price:   order.Price,
	}, nil
}
