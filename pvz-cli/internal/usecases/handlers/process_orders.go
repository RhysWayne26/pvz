package handlers

import (
	"context"
	"errors"
	"fmt"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
	"pvz-cli/internal/usecases/services"
)

// HandleProcessOrders processes orders for issue or return actions
func (f *DefaultFacadeHandler) HandleProcessOrders(
	ctx context.Context,
	req requests.ProcessOrdersRequest,
) (responses.ProcessOrdersResponse, error) {
	select {
	case <-ctx.Done():
		return responses.ProcessOrdersResponse{}, ctx.Err()
	default:
	}

	var results []services.ProcessResult

	switch req.Action {
	case constants.ActionIssue:
		results = f.orderService.IssueOrders(requests.IssueOrdersRequest{
			UserID:   req.UserID,
			OrderIDs: req.OrderIDs,
		})

	case constants.ActionReturn:
		results = f.orderService.CreateClientReturns(requests.ClientReturnsRequest{
			UserID:   req.UserID,
			OrderIDs: req.OrderIDs,
		})

	default:
		return responses.ProcessOrdersResponse{},
			apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", req.Action)
	}

	return buildProcessOrdersResponse(results), nil
}

func buildProcessOrdersResponse(results []services.ProcessResult) responses.ProcessOrdersResponse {
	resp := responses.ProcessOrdersResponse{
		Processed: make([]uint64, 0, len(results)),
		Failed:    make(map[uint64]responses.FailedOrder),
	}

	for _, r := range results {
		if r.Error == nil {
			resp.Processed = append(resp.Processed, r.OrderID)
			fmt.Printf("PROCESSED: %d\n", r.OrderID)
			continue
		}

		var appErr *apperrors.AppError
		if errors.As(r.Error, &appErr) {
			resp.Failed[r.OrderID] = responses.FailedOrder{
				Code:    appErr.Code,
				Message: appErr.Message,
			}
		} else {
			resp.Failed[r.OrderID] = responses.FailedOrder{
				Code:    apperrors.InternalError,
				Message: r.Error.Error(),
			}
		}
		apperrors.Handle(r.Error)
	}

	return resp
}
