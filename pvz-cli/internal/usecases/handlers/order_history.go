package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleOrderHistory returns all order history entries in a response model.
func (f *DefaultFacadeHandler) HandleOrderHistory(ctx context.Context, req requests.OrderHistoryRequest) (responses.OrderHistoryResponse, error) {
	select {
	case <-ctx.Done():
		return responses.OrderHistoryResponse{}, ctx.Err()
	default:
	}

	entries, err := f.historyService.ListAll(req.Page, req.Limit)
	if err != nil {
		return responses.OrderHistoryResponse{}, err
	}

	return responses.OrderHistoryResponse{
		History: entries,
	}, nil
}
