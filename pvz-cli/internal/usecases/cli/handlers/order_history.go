package handlers

import (
	"fmt"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/services"
)

// OrderHistoryHandler handles the order history command.
type OrderHistoryHandler struct {
	service services.HistoryService
}

// NewOrderHistoryHandler creates an instance of OrderHistoryHandler.
func NewOrderHistoryHandler(svc services.HistoryService) *OrderHistoryHandler {
	return &OrderHistoryHandler{
		service: svc,
	}
}

// Handle processes order-history command and displays all order events
func (h *OrderHistoryHandler) Handle() error {
	entries, err := h.service.ListAll(constants.DefaultHistoryPage, constants.DefaultHistoryLimit)
	if err != nil {
		return err
	}

	for _, e := range entries {
		fmt.Printf("HISTORY: %s %s %s\n", e.OrderID, e.Event, e.Timestamp.Format(constants.HistoryTimeLayout))
	}
	return nil
}
