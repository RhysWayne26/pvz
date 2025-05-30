package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

// ReturnOrderParams contains parameters for return-order command
type ReturnOrderParams struct {
	OrderID string `json:"order_id"`
}

// HandleReturnOrderCommand processes return-order command to return order to courier
func HandleReturnOrderCommand(params ReturnOrderParams, svc services.ReturnService) {
	orderID := strings.TrimSpace(params.OrderID)

	req := requests.ReturnOrderRequest{
		OrderID: orderID,
	}

	if err := svc.ReturnToCourier(req); err != nil {
		apperrors.Handle(err)
		return
	}

	fmt.Printf("ORDER_RETURNED: %s\n", orderID)
}
