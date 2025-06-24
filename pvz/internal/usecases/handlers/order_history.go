package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleOrderHistory returns all order history entries in a response model.
func (f *DefaultFacadeHandler) HandleOrderHistory(ctx context.Context, req requests.OrderHistoryFilter) (responses.OrderHistoryResponse, error) {
	if ctx.Err() != nil {
		return responses.OrderHistoryResponse{}, ctx.Err()
	}

	entries, err := f.historyService.List(ctx, req)
	if err != nil {
		return responses.OrderHistoryResponse{}, err
	}

	return responses.OrderHistoryResponse{
		History: entries,
	}, nil
}
