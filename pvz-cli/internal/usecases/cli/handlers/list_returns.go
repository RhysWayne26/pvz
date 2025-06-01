package handlers

import (
	"fmt"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
)

// ListReturnsHandler handles the list returns command.
type ListReturnsHandler struct {
	params  dto.ListReturnsParams
	service services.ReturnService
}

// NewListReturnsHandler creates an instance of ListReturnsHandler.
func NewListReturnsHandler(p dto.ListReturnsParams, svc services.ReturnService) *ListReturnsHandler {
	return &ListReturnsHandler{
		params:  p,
		service: svc,
	}
}

// Handle processes list-returns command with pagination
func (h *ListReturnsHandler) Handle() error {
	page := constants.DefaultPage
	limit := constants.DefaultLimit
	if h.params.Page != nil {
		page = *h.params.Page
	}
	if h.params.Limit != nil {
		limit = *h.params.Limit
	}

	if err := utils.ValidatePositiveInt("page", h.params.Page); err != nil {
		return err
	}

	if err := utils.ValidatePositiveInt("limit", h.params.Limit); err != nil {
		return err
	}

	entries, err := h.service.ListReturns(page, limit)
	if err != nil {
		return err
	}

	for _, r := range entries {
		fmt.Printf("RETURN: %s %s %s\n", r.OrderID, r.UserID, r.ReturnedAt.Format(constants.TimeLayout))
	}
	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
	return nil
}
