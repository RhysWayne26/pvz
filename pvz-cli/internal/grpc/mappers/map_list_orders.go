package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbListOrdersRequest maps a gRPC OrdersFilterRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbListOrdersRequest(in *pb.ListOrdersRequest) requests.OrdersFilterRequest {
	req := requests.OrdersFilterRequest{
		UserID: &in.UserId,
		InPvz:  &in.InPvz,
	}

	if in.LastN != nil {
		last := int(*in.LastN)
		req.Last = &last
	}

	if in.Pagination != nil {
		page := int(in.Pagination.Page)
		limit := int(in.Pagination.CountOnPage)
		req.Page = &page
		req.Limit = &limit
	}

	return req
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
