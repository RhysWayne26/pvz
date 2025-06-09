package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleAcceptOrder processes accept-order command with package pricing validation, optionally suppressing output for batch import
func (f *DefaultFacadeHandler) HandleAcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (responses.AcceptOrderResponse, error) {
	select {
	case <-ctx.Done():
		return responses.AcceptOrderResponse{}, ctx.Err()
	default:
	}

	order, err := f.orderService.AcceptOrder(req)
	if err != nil {
		return responses.AcceptOrderResponse{}, err
	}

	return responses.AcceptOrderResponse{
		OrderID: order.OrderID,
		Package: order.Package,
		Price:   order.Price,
	}, nil
}
