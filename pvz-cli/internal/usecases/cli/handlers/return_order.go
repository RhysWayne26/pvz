package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
)

type ReturnOrderParams struct {
	OrderID string `json:"orderID"`
}

func HandleReturnOrderCommand(params ReturnOrderParams, svc services.ReturnService) {
	orderID, err := uuid.Parse(params.OrderID)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid order_id"))
		return
	}

	req := requests.ReturnOrderRequest{
		OrderID: orderID,
	}

	if err := svc.ReturnToCourier(req); err != nil {
		apperrors.Handle(err)
		return
	}

	fmt.Printf("ORDER_RETURNED: %s\n", orderID)
}
