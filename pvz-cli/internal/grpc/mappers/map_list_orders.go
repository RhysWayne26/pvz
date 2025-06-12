package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbListOrdersRequest maps a gRPC OrdersFilterRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbListOrdersRequest(in *pb.ListOrdersRequest) (requests.OrdersFilterRequest, error) {
	if err := providedUserIDCheck(in.UserId); err != nil {
		return requests.OrdersFilterRequest{}, err
	}
	opts := []requests.FilterOption{
		requests.WithUserID(in.UserId),
		requests.WithInPvz(in.InPvz),
	}
	if in.LastN != nil {
		opts = append(opts, requests.WithLast(int(*in.LastN)))
	}

	opts = append(opts, collectPaginationOptions(in.Pagination)...)
	return requests.NewOrdersFilter(opts...), nil
}

// ToPbOrdersList maps the internal ListOrdersResponse to a gRPC OrdersList response.
func (f *DefaultGRPCFacadeMapper) ToPbOrdersList(res responses.ListOrdersResponse) *pb.OrdersList {
	orders := res.Orders
	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		pbOrders = append(pbOrders, toPbOrder(o))
	}

	var totalValue int32
	if res.Total != nil {
		totalValue = int32(*res.Total)
	}

	return &pb.OrdersList{
		Orders: pbOrders,
		Total:  totalValue,
	}
}
