package mappers

import (
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/common/utils"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbAcceptOrderRequest converts a gRPC AcceptOrderRequest into an internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbAcceptOrderRequest(in *pb.AcceptOrderRequest) (requests.AcceptOrderRequest, error) {
	var pkg models.PackageType
	if in.Package != nil {
		pkg = fromPbPackageType(*in.Package)
	}

	if err := utils.ValidateFractionDigits("weight", in.Weight, constants.WeightFractionDigit); err != nil {
		return requests.AcceptOrderRequest{}, err
	}
	if err := utils.ValidateFractionDigits("price", in.Price, constants.PriceFractionDigit); err != nil {
		return requests.AcceptOrderRequest{}, err
	}

	return requests.AcceptOrderRequest{
		OrderID:   in.OrderId,
		UserID:    in.UserId,
		ExpiresAt: in.ExpiresAt.AsTime(),
		Weight:    in.Weight,
		Price:     in.Price,
		Package:   pkg,
	}, nil
}

// ToPbAcceptOrderResponse converts an internal AcceptOrderResponse into a gRPC OrderResponse.
func (f *DefaultGRPCFacadeMapper) ToPbAcceptOrderResponse(res responses.AcceptOrderResponse) *pb.OrderResponse {
	return &pb.OrderResponse{
		OrderId: res.OrderID,
		Status:  pb.OrderStatus_ORDER_STATUS_ACCEPTED,
	}
}
