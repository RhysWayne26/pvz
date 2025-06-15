package mappers

import (
	"math"
	"pvz-cli/internal/common/apperrors"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbImportOrdersRequest maps gRPC ImportOrdersRequest to internal model.
func (f *DefaultGRPCFacadeMapper) FromPbImportOrdersRequest(
	in *pb.ImportOrdersRequest,
) requests.ImportOrdersRequest {
	statuses := make([]requests.ImportOrderStatus, 0, len(in.Orders))

	for i, o := range in.Orders {
		itemNumber := i + 1
		orderID := o.OrderId

		req, err := f.FromPbAcceptOrderRequest(o)
		statuses = append(statuses, requests.ImportOrderStatus{
			ItemNumber: itemNumber,
			OrderID:    orderID,
			Request:    &req,
			Error:      err,
		})
	}

	return requests.ImportOrdersRequest{
		Statuses: statuses,
	}
}

// ToPbImportResult maps internal ImportOrdersResponse to protobuf ImportResult.
func (f *DefaultGRPCFacadeMapper) ToPbImportResult(res responses.ImportOrdersResponse) *pb.ImportResult {
	fails := make([]*pb.FailedBatchedOrder, 0, len(res.Statuses))
	for _, status := range res.Statuses {
		if status.Error != nil {
			fails = append(fails, &pb.FailedBatchedOrder{
				OrderId: status.OrderID,
				Code:    apperrors.CodeFromError(status.Error),
				Reason:  apperrors.MessageFromError(status.Error),
			})
		}
	}
	var imported int32
	if res.Imported > math.MaxInt32 {
		imported = math.MaxInt32
	} else {
		imported = int32(res.Imported) // #nosec G115
	}
	return &pb.ImportResult{
		Imported: imported,
		Errors:   fails,
	}
}
