package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
)

type ListOrdersParams struct {
	UserID   string
	InPvz    bool
	UseInPvz bool
	Last     *int
	LastID   string
	Page     *int
	Limit    *int
}

func HandleListOrdersCommand(params ListOrdersParams, svc services.OrderService) {
	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid user_id"))
		return
	}

	var lastID *uuid.UUID
	if params.LastID != "" {
		parsed, err := uuid.Parse(params.LastID)
		if err != nil {
			apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid last_id"))
			return
		}
		lastID = &parsed
	}

	var inPvzPtr *bool
	if params.UseInPvz {
		inPvzPtr = &params.InPvz
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
		fmt.Printf("ORDER: %s %s %s %s\n", o.OrderID, o.UserID, o.Status, o.ExpiresAt.Format(constants.TimeLayout))
	}
	fmt.Printf("TOTAL: %d\n", total)
}
