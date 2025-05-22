package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"time"
)

type AcceptOrderParams struct {
	OrderID   string `json:"orderID"`
	UserID    string `json:"userID"`
	ExpiresAt string `json:"expiresAt"`
}

func HandleAcceptOrderCommand(params AcceptOrderParams, svc services.OrderService) {
	orderID, err := uuid.Parse(params.OrderID)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid order_id"))
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid user_id"))
		return
	}

	expiresAt, err := time.Parse(constants.TimeLayout, params.ExpiresAt)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid expires_at format"))
		return
	}

	req := requests.AcceptOrderRequest{
		OrderID:   orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	if err := svc.AcceptOrder(req); err != nil {
		apperrors.Handle(err)
		return
	}

	fmt.Printf("ORDER_ACCEPTED: %s\n", orderID)
}
