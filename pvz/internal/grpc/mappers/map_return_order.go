package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbReturnOrderRequest maps a gRPC OrderIdRequest to the internal ReturnOrderRequest.
func (f *DefaultGRPCFacadeMapper) FromPbReturnOrderRequest(in *pb.OrderIdRequest) (requests.ReturnOrderRequest, error) {
	if err := providedOrderIDCheck(in.OrderId); err != nil {
		return requests.ReturnOrderRequest{}, err
	}

	return requests.ReturnOrderRequest{
		OrderID: in.OrderId,
	}, nil
}

// ToPbReturnOrderResponse maps the internal ReturnOrderResponse to a gRPC OrderResponse.
func (f *DefaultGRPCFacadeMapper) ToPbReturnOrderResponse(res responses.ReturnOrderResponse) *pb.OrderResponse {
	return &pb.OrderResponse{
		OrderId: res.OrderID,
		Status:  pb.OrderStatus_ORDER_STATUS_RETURNED_TO_WAREHOUSE,
	}
}
