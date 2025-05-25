package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
	"time"
)

type AcceptOrderParams struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ExpiresAt string `json:"expires_at"`
}

func HandleAcceptOrderCommand(params AcceptOrderParams, svc services.OrderService) {
	orderID := strings.TrimSpace(params.OrderID)
	userID := strings.TrimSpace(params.UserID)

	expiresAt, err := time.Parse(constants.TimeLayout, strings.TrimSpace(params.ExpiresAt))
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
