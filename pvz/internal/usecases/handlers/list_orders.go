package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
	"time"
)

const listCacheTTL = 5 * time.Minute

// HandleListOrders processes the list-orders request and returns the result.
func (f *DefaultFacadeHandler) HandleListOrders(ctx context.Context, req requests.OrdersFilterRequest) (responses.ListOrdersResponse, error) {
	if ctx.Err() != nil {
		return responses.ListOrdersResponse{}, ctx.Err()
	}
	key := fmt.Sprintf(
		"ListOrders:user=%v;inPvz=%v;page=%v;limit=%v",
		req.UserID, req.InPvz, req.Page, req.Limit,
	)
	if raw, ok := f.responsesCache.Get(key); ok {
		if cached, ok2 := raw.(responses.ListOrdersResponse); ok2 {
			f.metrics.IncOrdersServed(float64(len(cached.Orders)))
			return cached, nil
		}
	}
	orders, nextID, total, err := f.orderService.ListOrders(ctx, req)
	if err != nil {
		return responses.ListOrdersResponse{}, err
	}
	resp := responses.ListOrdersResponse{
		Orders: orders,
		NextID: &nextID,
		Total:  utils.Ptr(total),
	}
	f.responsesCache.Set(key, resp, listCacheTTL)
	f.metrics.IncOrdersServed(float64(len(resp.Orders)))
	return resp, nil
}
