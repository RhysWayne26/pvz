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

// ProcessOrdersHandler handles the process orders command.
type ProcessOrdersHandler struct {
	params  dto.ProcessOrdersParams
	service services.OrderService
}

// NewProcessOrdersHandler creates an instance of ProcessOrdersHandler.
func NewProcessOrdersHandler(p dto.ProcessOrdersParams, orderSvc services.OrderService) *ProcessOrdersHandler {
	return &ProcessOrdersHandler{
		params:  p,
		service: orderSvc,
	}
}

// Handle processes orders for issue or return actions
func (h *ProcessOrdersHandler) Handle() error {
	id := strings.TrimSpace(h.params.UserID)
	orderIDs := strings.Split(h.params.OrderIDs, ",")
	if len(orderIDs) == 0 {
		return apperrors.Newf(apperrors.ValidationFailed, "no order IDs provided")
	}

	switch h.params.Action {
	case constants.ActionIssue:
		req := requests.IssueOrdersRequest{UserID: id, OrderIDs: orderIDs}
		results := h.service.IssueOrders(req)

		for _, res := range results {
			if res.Error == nil {
				fmt.Printf("PROCESSED: %s\n", res.OrderID)
			} else {
				apperrors.Handle(res.Error)
			}
		}

	case constants.ActionReturn:
		req := requests.ClientReturnsRequest{UserID: id, OrderIDs: orderIDs}
		results := h.service.CreateClientReturns(req)

		for _, res := range results {
			if res.Error == nil {
				fmt.Printf("PROCESSED: %s\n", res.OrderID)
			} else {
				apperrors.Handle(res.Error)
			}
		}

	default:
		return apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", h.params.Action)
	}
	return nil
}
