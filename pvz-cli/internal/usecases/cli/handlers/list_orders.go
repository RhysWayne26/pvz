package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
	"strings"
)

// HandleListOrdersCommand processes list-orders command with filtering and pagination
func HandleListOrdersCommand(params dto.ListOrdersParams, svc services.OrderService) {
	userID := strings.TrimSpace(params.UserID)
	var lastID string
	if params.LastID != "" {
		parsed := strings.TrimSpace(params.LastID)
		lastID = parsed
	}

	var inPvzPtr *bool
	if params.InPvz != nil {
		inPvzPtr = params.InPvz
	}

	if err := utils.ValidatePositiveInt("last", params.Last); err != nil {
		apperrors.Handle(err)
		return
	}

	if err := utils.ValidatePositiveInt("page", params.Page); err != nil {
		apperrors.Handle(err)
		return
	}

	if err := utils.ValidatePositiveInt("limit", params.Limit); err != nil {
		apperrors.Handle(err)
		return
	}

	filter := requests.ListOrdersFilter{
		UserID: userID,
		InPvz:  inPvzPtr,
		LastID: lastID,
		Page:   params.Page,
		Limit:  params.Limit,
		Last:   params.Last,
	}

	orders, _, total, err := svc.ListOrders(filter)
	if err != nil {
		apperrors.Handle(err)
		return
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
}
