package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleListReturns processes list-returns command with pagination
func (f *DefaultFacadeHandler) HandleListReturns(ctx context.Context, req requests.ListReturnsRequest) (responses.ListReturnsResponse, error) {
	select {
	case <-ctx.Done():
		return responses.ListReturnsResponse{}, ctx.Err()
	default:
	}

	entries, err := f.orderService.ListReturns(req.Page, req.Limit)
	if err != nil {
		return responses.ListReturnsResponse{}, err
	}

	return responses.ListReturnsResponse{
		Returns: entries,
		Page:    req.Page,
		Limit:   req.Limit,
	}, nil
}
