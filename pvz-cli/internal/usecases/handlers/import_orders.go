package handlers

import (
	"context"
	"errors"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

const silentAcceptOrderOutput = true

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

	var (
		importedCount int32
		failures      = make(map[uint64]responses.FailedImport)
	)

	for _, order := range req.Orders {
		_, err := f.HandleAcceptOrder(ctx, order, silentAcceptOrderOutput)
		if err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				failures[order.OrderID] = responses.FailedImport{
					Code:    appErr.Code,
					Message: appErr.Message,
				}
			} else {
				failures[order.OrderID] = responses.FailedImport{
					Code:    apperrors.InternalError,
					Message: err.Error(),
				}
			}
			continue
		}
		importedCount++
	}

	return responses.ImportOrdersResponse{
		Imported: importedCount,
		Errors:   failures,
	}, nil
}
