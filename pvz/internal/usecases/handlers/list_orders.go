package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
	"time"
)

const listCacheTTL = 30 * time.Second

// HandleListOrders processes the list-orders request and returns the result.
func (f *DefaultFacadeHandler) HandleListOrders(ctx context.Context, req requests.OrdersFilterRequest) (responses.ListOrdersResponse, error) {
	if ctx.Err() != nil {
		return responses.ListOrdersResponse{}, ctx.Err()
	}
	uid := uint64(0)
	if req.UserID != nil {
		uid = *req.UserID
	}
	inPvz := false
	if req.InPvz != nil {
		inPvz = *req.InPvz
	}
	page := constants.DefaultPage
	if req.Page != nil {
		page = *req.Page
	}
	limit := constants.DefaultLimit
	if req.Limit != nil {
		limit = *req.Limit
	}
	key := fmt.Sprintf(
		"ListOrders:user=%d;inPvz=%t;page=%d;limit=%d",
		uid, inPvz, page, limit,
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
