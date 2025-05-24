package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/utils"
	"strings"
)

type ProcessOrdersParams struct {
	UserID   string `json:"user_id"`
	Action   string `json:"action"`
	OrderIDs string `json:"order_ids"`
}

func HandleProcessOrders(params ProcessOrdersParams,
	orderSvc services.OrderService, returnSvc services.ReturnService,
) {
	uid := strings.TrimSpace(params.UserID)
	rawOrders := strings.Split(params.OrderIDs, ",")
	ids := utils.UniqueStrings(rawOrders)
	if len(ids) == 0 {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided"))
		return
	}

	switch params.Action {
	case constants.ActionIssue:
		req := requests.IssueOrderRequest{UserID: uid, OrderIDs: ids}
		if err := orderSvc.IssueOrder(req); err != nil {
			apperrors.Handle(err)
			return
		}
		for _, id := range ids {
			fmt.Printf("PROCESSED: %s\n", id)
		}

	case constants.ActionReturn:
		req := requests.ClientReturnRequest{UserID: uid, OrderIDs: ids}
		if err := returnSvc.CreateClientReturn(req); err != nil {
			apperrors.Handle(err)
			return
		}
		for _, id := range ids {
			fmt.Printf("PROCESSED: %s\n", id)
		}

	default:
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", params.Action))
	}
}
