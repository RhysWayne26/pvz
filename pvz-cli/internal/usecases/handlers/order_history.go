package handlers

import (
	"context"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/usecases/responses"
)

// HandleOrderHistory returns all order history entries in a response model.
func (f *DefaultFacadeHandler) HandleOrderHistory(ctx context.Context) (responses.OrderHistoryResponse, error) {
	select {
	case <-ctx.Done():
		return responses.OrderHistoryResponse{}, ctx.Err()
	default:
	}

	entries, err := f.historyService.ListAll(constants.DefaultHistoryPage, constants.DefaultHistoryLimit)
	if err != nil {
		return responses.OrderHistoryResponse{}, err
	}

	return responses.OrderHistoryResponse{
		History: entries,
	}, nil
}
