package handlers

import (
	"fmt"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
	"strings"
)

// ListOrdersHandler handles the list order command.
type ListOrdersHandler struct {
	params  dto.ListOrdersParams
	service services.OrderService
}

// NewListOrdersHandler creates an instance of ListOrdersHandler.
func NewListOrdersHandler(p dto.ListOrdersParams, svc services.OrderService) *ListOrdersHandler {
	return &ListOrdersHandler{
		params:  p,
		service: svc,
	}

}

// Handle processes list-orders command with filtering and pagination.
func (h *ListOrdersHandler) Handle() error {
	userID := strings.TrimSpace(h.params.UserID)
	var lastID string
	if h.params.LastID != "" {
		parsed := strings.TrimSpace(h.params.LastID)
		lastID = parsed
	}

	var inPvzPtr *bool
	if h.params.InPvz != nil {
		inPvzPtr = h.params.InPvz
	}

	if err := utils.ValidatePositiveInt("last", h.params.Last); err != nil {
		return err
	}

	if err := utils.ValidatePositiveInt("page", h.params.Page); err != nil {
		return err
	}

	if err := utils.ValidatePositiveInt("limit", h.params.Limit); err != nil {
		return err
	}

	filter := requests.ListOrdersFilter{
		UserID: userID,
		InPvz:  inPvzPtr,
		LastID: lastID,
		Page:   h.params.Page,
		Limit:  h.params.Limit,
		Last:   h.params.Last,
	}

	orders, _, total, err := h.service.ListOrders(filter)
	if err != nil {
		return err
	}

	for _, o := range orders {
		fmt.Printf("ORDER: %s %s %s %s %s %.*f %.*f\n",
			o.OrderID,
			o.UserID,
			o.Status,
			o.ExpiresAt.Format(constants.TimeLayout),
			o.Package,
			constants.WeightFractionDigit, o.Weight,
			constants.PriceFractionDigit, o.Price,
		)
	}
	fmt.Printf("TOTAL: %d\n", total)
	return nil
}
