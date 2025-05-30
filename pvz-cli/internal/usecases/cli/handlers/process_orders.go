package handlers

import (
	"fmt"
	"pvz-cli/internal/apperrors"
	"pvz-cli/internal/constants"
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services"
	"strings"
)

// HandleProcessOrders processes orders for issue or return actions
func HandleProcessOrders(
	params dto.ProcessOrdersParams,
	orderSvc services.OrderService,
	returnSvc services.ReturnService,
) {
	id := strings.TrimSpace(params.UserID)
	orderIDs := strings.Split(params.OrderIDs, ",")
	if len(orderIDs) == 0 {
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided"))
		return
	}

	switch params.Action {
	case constants.ActionIssue:
		req := requests.IssueOrdersRequest{UserID: id, OrderIDs: orderIDs}
		results := orderSvc.IssueOrders(req)

		for _, res := range results {
			if res.Error == nil {
				fmt.Printf("PROCESSED: %s\n", res.OrderID)
			} else {
				apperrors.Handle(res.Error)
			}
		}

	case constants.ActionReturn:
		req := requests.ClientReturnsRequest{UserID: id, OrderIDs: orderIDs}
		results := returnSvc.CreateClientReturns(req)

		for _, res := range results {
			if res.Error == nil {
				fmt.Printf("PROCESSED: %s\n", res.OrderID)
			} else {
				apperrors.Handle(res.Error)
			}
		}

	default:
		apperrors.Handle(apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", params.Action))
	}
}
