package handlers

import (
	"context"
	"fmt"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleImportOrders executes the import-orders command by invoking HandleAcceptOrder for each parsed order.
// It returns an ImportOrdersResponse containing the count of successfully imported orders and a map of failures.
func (f *DefaultFacadeHandler) HandleImportOrders(
	ctx context.Context,
	req requests.ImportOrdersRequest,
) (responses.ImportOrdersResponse, error) {
	if err := ctx.Err(); err != nil {
		return responses.ImportOrdersResponse{}, err
	}
	batchResults, err := f.orderService.ImportOrders(ctx, req)
	if err != nil {
		return responses.ImportOrdersResponse{}, err
	}
	statuses := make([]requests.ImportOrderStatus, len(req.Statuses))
	copy(statuses, req.Statuses)
	importedCount := 0
	for i, r := range batchResults {
		statuses[i].Error = r.Error
		if r.Error == nil {
			importedCount++
			f.responsesCache.Invalidate(fmt.Sprintf("OrderHistory:%d", r.OrderID))
		}
	}
	if importedCount > 0 {
		f.responsesCache.InvalidatePattern("^ListOrders:")
		f.metrics.IncOrdersServed(float64(importedCount))
	}
	return responses.ImportOrdersResponse{
		Imported: importedCount,
		Statuses: statuses,
	}, nil
}
