package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbAcceptOrderRequest converts a gRPC AcceptOrderRequest into an internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbAcceptOrderRequest(in *pb.AcceptOrderRequest) requests.AcceptOrderRequest {
	var pkg models.PackageType
	if in.Package != nil {
		pkg = fromPbPackageType(*in.Package)
	}

	return requests.AcceptOrderRequest{
		OrderID:   in.OrderId,
		UserID:    in.UserId,
		ExpiresAt: in.ExpiresAt.AsTime(),
		Weight:    float64(in.Weight),
		Price:     float64(in.Price),
		Package:   pkg,
	}
}

// ToPbAcceptOrderResponse converts an internal AcceptOrderResponse into a gRPC OrderResponse.
func (f *DefaultGRPCFacadeMapper) ToPbAcceptOrderResponse(res responses.AcceptOrderResponse) *pb.OrderResponse {
	return &pb.OrderResponse{
		OrderId: res.OrderID,
		Status:  pb.OrderStatus_ORDER_STATUS_ACCEPTED,
	}
}
