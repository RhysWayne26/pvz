package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbImportOrdersRequest maps protobuf ImportOrdersRequest to internal DTO.
func (f *DefaultGRPCFacadeMapper) FromPbImportOrdersRequest(in *pb.ImportOrdersRequest) requests.ImportOrdersRequest {
	orders := make([]requests.AcceptOrderRequest, 0, len(in.Orders))
	for _, o := range in.Orders {
		orders = append(orders, f.FromPbAcceptOrderRequest(o))
	}
	return requests.ImportOrdersRequest{
		Orders: orders,
	}
}

// ToPbImportResult maps internal ImportOrdersResponse to protobuf ImportResult.
func (f *DefaultGRPCFacadeMapper) ToPbImportResult(res responses.ImportOrdersResponse) *pb.ImportResult {
	errors := make([]*pb.FailedImport, 0, len(res.Errors))
	for orderID, err := range res.Errors {
		errors = append(errors, &pb.FailedImport{
			OrderId: orderID,
			Code:    string(err.Code),
			Reason:  err.Message,
		})
	}
	return &pb.ImportResult{
		Imported: res.Imported,
		Errors:   errors,
	}
}
