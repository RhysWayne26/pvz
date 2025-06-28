package handlers

import (
	"context"
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
	if ctx.Err() != nil {
		return responses.ProcessOrdersResponse{}, ctx.Err()
	}

	var results []services.ProcessResult
	var err error

	switch req.Action {
	case constants.ActionIssue:
		results, err = f.OrderService.IssueOrders(ctx,
			requests.IssueOrdersRequest{
				UserID:   req.UserID,
				OrderIDs: req.OrderIDs,
			})

	case constants.ActionReturn:
		results, err = f.OrderService.CreateClientReturns(ctx,
			requests.ClientReturnsRequest{
				UserID:   req.UserID,
				OrderIDs: req.OrderIDs,
			})

	default:
		return responses.ProcessOrdersResponse{},
			apperrors.Newf(apperrors.ValidationFailed, "unknown action %q", req.Action)
	}

	if err != nil {
		return responses.ProcessOrdersResponse{}, err
	}

	return buildProcessOrdersResponse(results), nil
}

func buildProcessOrdersResponse(resultsFromService []services.ProcessResult) responses.ProcessOrdersResponse {
	res := responses.ProcessOrdersResponse{
		Processed: make([]uint64, 0, len(resultsFromService)),
		Failed:    make([]responses.ProcessFailReport, 0, len(resultsFromService)),
	}

	for _, result := range resultsFromService {
		if result.Error != nil {
			res.Failed = append(res.Failed, responses.ProcessFailReport{
				OrderID: result.OrderID,
				Error:   result.Error,
			})
		} else {
			res.Processed = append(res.Processed, result.OrderID)
		}

	}

	return res
}
