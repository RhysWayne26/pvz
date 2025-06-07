package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// FromPbListReturnsRequest maps a gRPC ListReturnsRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbListReturnsRequest(in *pb.ListReturnsRequest) requests.ListReturnsRequest {
	var req requests.ListReturnsRequest
	if in.Pagination != nil {
		page := int(in.Pagination.Page)
		limit := int(in.Pagination.CountOnPage)
		req.Page = page
		req.Limit = limit
	}

	return req
}

// ToPbReturnsList maps the internal ListReturnsResponse to a gRPC ReturnsList response.
func (f *DefaultGRPCFacadeMapper) ToPbReturnsList(res responses.ListReturnsResponse) *pb.ReturnsList {
	pbOrders := make([]*pb.Order, 0, len(res.Returns))
	for _, r := range res.Returns {
		pbOrders = append(pbOrders, &pb.Order{
			OrderId:   r.OrderID,
			UserId:    r.UserID,
			Status:    pb.OrderStatus_ORDER_STATUS_RETURNED,
			ExpiresAt: timestamppb.New(r.ReturnedAt),
		})
	}
	return &pb.ReturnsList{
		Returns: pbOrders,
	}
}
