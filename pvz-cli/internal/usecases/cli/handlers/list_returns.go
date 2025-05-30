package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
)

// HandleListReturnsCommand processes list-returns command with pagination
func HandleListReturnsCommand(params dto.ListReturnsParams, svc services.ReturnService) {
	page := constants.DefaultPage
	limit := constants.DefaultLimit
	if params.Page != nil {
		page = *params.Page
	}
	if params.Limit != nil {
		limit = *params.Limit
	}

	if err := utils.ValidatePositiveInt("page", params.Page); err != nil {
		apperrors.Handle(err)
		return
	}

	if err := utils.ValidatePositiveInt("limit", params.Limit); err != nil {
		apperrors.Handle(err)
		return
	}

	entries, err := svc.ListReturns(page, limit)
	if err != nil {
		apperrors.Handle(err)
		return
	}

	for _, r := range entries {
		fmt.Printf("RETURN: %s %s %s\n", r.OrderID, r.UserID, r.ReturnedAt.Format(constants.TimeLayout))
	}
	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
}
