package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

type ReturnOrderParams struct {
	OrderID string `json:"order_id"`
}

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
