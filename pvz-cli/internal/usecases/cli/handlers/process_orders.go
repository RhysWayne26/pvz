package handlers

import (
	"fmt"
	"github.com/google/uuid"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

type ProcessOrdersParams struct {
	UserID   string `json:"userID"`
	Action   string `json:"action"`
	OrderIDs string `json:"orderIDs"`
}

func HandleProcessOrders(params ProcessOrdersParams,
	orderSvc services.OrderService, returnSvc services.ReturnService,
) {
	uid, err := uuid.Parse(params.UserID)
	if err != nil {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid user_id"))
		return
	}

	parts := strings.Split(params.OrderIDs, ",")
	var ids []uuid.UUID
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, e := uuid.Parse(p)
		if e != nil {
			apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "invalid order_id: %s", p))
			return
		}
		ids = append(ids, id)
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
