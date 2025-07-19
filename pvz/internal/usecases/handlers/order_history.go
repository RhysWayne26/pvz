package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
	"time"
)

const historyCacheTTL = 5 * time.Minute

// HandleOrderHistory returns all order history entries in a response model.
func (f *DefaultFacadeHandler) HandleOrderHistory(ctx context.Context, req requests.OrderHistoryFilter) (responses.OrderHistoryResponse, error) {
	if ctx.Err() != nil {
		return responses.OrderHistoryResponse{}, ctx.Err()
	}
	if req.OrderID != nil {
		key := fmt.Sprintf("OrderHistory:%d", *req.OrderID)
		if raw, ok := f.responsesCache.Get(key); ok {
			if resp, ok := raw.(responses.OrderHistoryResponse); ok {
				return resp, nil
			}
		}
		entries, err := f.historyService.List(ctx, req)
		if err != nil {
			return responses.OrderHistoryResponse{}, err
		}

		resp := responses.OrderHistoryResponse{History: entries}
		f.responsesCache.Set(key, resp, historyCacheTTL)
		return resp, nil
	}
	entries, err := f.historyService.List(ctx, req)
	if err != nil {
		return responses.OrderHistoryResponse{}, err
	}

	return responses.OrderHistoryResponse{
		History: entries,
	}, nil
}
