package mappers

import (
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

		req, err := f.FromPbAcceptOrderRequest(o)
		statuses = append(statuses, requests.ImportOrderStatus{
			ItemNumber: itemNumber,
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
	fails := make([]*pb.FailedImport, 0, len(res.Statuses))
	for _, status := range res.Statuses {
		if status.Error != nil {
			fails = append(fails, &pb.FailedImport{
				OrderId: status.Request.OrderID,
				Code:    apperrors.CodeFromError(status.Error),
				Reason:  status.Error.Error(),
			})
		}
	}
	return &pb.ImportResult{
		Imported: int32(res.Imported),
		Errors:   fails,
	}
}
