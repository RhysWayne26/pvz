package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// HandleImportOrders executes the import-orders command by invoking HandleAcceptOrder for each parsed order.
// It returns an ImportOrdersResponse containing the count of successfully imported orders and a map of failures.
func (f *DefaultFacadeHandler) HandleImportOrders(
	ctx context.Context,
	req requests.ImportOrdersRequest,
) (responses.ImportOrdersResponse, error) {
	select {
	case <-ctx.Done():
		return responses.ImportOrdersResponse{}, ctx.Err()
	default:
	}

	var importedCount int
	for i := range req.Statuses {
		status := &req.Statuses[i]

		if status.Error != nil {
			continue
		}
		_, err := f.HandleAcceptOrder(ctx, *status.Request)
		if err != nil {
			status.Error = err
			continue
		}
		importedCount++
	}

	return responses.ImportOrdersResponse{
		Imported: importedCount,
		Statuses: req.Statuses,
	}, nil
}
