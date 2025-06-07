package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleAcceptOrder processes accept-order command with package pricing validation, optionally suppressing output for batch import
func (f *DefaultFacadeHandler) HandleAcceptOrder(ctx context.Context, req requests.AcceptOrderRequest, silent bool) (responses.AcceptOrderResponse, error) {
	select {
	case <-ctx.Done():
		return responses.AcceptOrderResponse{}, ctx.Err()
	default:
	}

	order, err := f.orderService.AcceptOrder(req)
	if err != nil {
		return responses.AcceptOrderResponse{}, err
	}

	if !silent {
		fmt.Printf("ORDER_ACCEPTED: %d\n", order.OrderID)
		fmt.Printf("PACKAGE: %s\n", order.Package)
		fmt.Printf("TOTAL_PRICE: %.*f\n", constants.PriceFractionDigit, order.Price)
	}

	return responses.AcceptOrderResponse{
		OrderID: order.OrderID,
		Package: order.Package,
		Price:   order.Price,
	}, nil
}
