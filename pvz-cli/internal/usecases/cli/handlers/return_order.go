package handlers

import (
	"fmt"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

// ReturnOrderHandler handles the return order command.
type ReturnOrderHandler struct {
	params  dto.ReturnOrderParams
	service services.OrderService
}

// NewReturnOrderHandler creates an instance of ReturnOrderHandler.
func NewReturnOrderHandler(p dto.ReturnOrderParams, svc services.OrderService) *ReturnOrderHandler {
	return &ReturnOrderHandler{
		params:  p,
		service: svc,
	}
}

// Handle processes return-order command to return order to courier
func (h *ReturnOrderHandler) Handle() error {
	orderID := strings.TrimSpace(h.params.OrderID)

	req := requests.ReturnOrderRequest{
		OrderID: orderID,
	}

	if err := h.service.ReturnToCourier(req); err != nil {
		return err
	}

	fmt.Printf("ORDER_RETURNED: %s\n", orderID)
	return nil
}
