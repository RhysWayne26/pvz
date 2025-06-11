package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbListReturnsRequest maps a gRPC ListReturnsRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbListReturnsRequest(in *pb.ListReturnsRequest) requests.OrdersFilterRequest {
	var req requests.OrdersFilterRequest
	if in.Pagination != nil {
		page := int(in.Pagination.Page)
		limit := int(in.Pagination.CountOnPage)
		status := models.Returned
		req.Page = &page
		req.Limit = &limit
		req.Status = &status
	}

	return req
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
