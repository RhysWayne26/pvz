package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbListReturnsRequest maps a gRPC ListReturnsRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbListReturnsRequest(in *pb.ListReturnsRequest) requests.OrdersFilterRequest {
	opts := []requests.FilterOption{
		requests.WithStatus(models.Returned),
	}
	opts = append(opts, collectPaginationOptions(in.Pagination)...)
	return requests.NewOrdersFilter(opts...)
}

// ToPbReturnsList maps the internal ListReturnsResponse to a gRPC ReturnsList response.
func (f *DefaultGRPCFacadeMapper) ToPbReturnsList(res responses.ListOrdersResponse) *pb.ReturnsList {
	returns := res.Orders
	pbOrders := make([]*pb.Order, 0, len(returns))
	for _, r := range returns {
		pbOrders = append(pbOrders, toPbOrder(r))
	}
	return &pb.ReturnsList{
		Returns: pbOrders,
	}
}
