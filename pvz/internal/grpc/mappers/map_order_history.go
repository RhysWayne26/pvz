package mappers

import (
	"pvz-cli/internal/common/constants"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// FromPbOrderHistoryRequest maps a gRPC GetHistoryRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbOrderHistoryRequest(in *pb.GetHistoryRequest) requests.OrderHistoryRequest {
	req := requests.OrderHistoryRequest{
		Page:  constants.DefaultHistoryPage,
		Limit: constants.DefaultHistoryLimit,
	}
	if in.Pagination != nil {
		if page := int(in.Pagination.Page); page > 0 {
			req.Page = page
		}
		if limit := int(in.Pagination.CountOnPage); limit > 0 {
			req.Limit = limit
		}
	}
	return req
}

// ToPbOrderHistoryList maps internal OrderHistoryResponse to protobuf OrderHistoryList.
func (f *DefaultGRPCFacadeMapper) ToPbOrderHistoryList(res responses.OrderHistoryResponse) *pb.OrderHistoryList {
	history := make([]*pb.OrderHistory, 0, len(res.History))
	for _, e := range res.History {
		history = append(history, &pb.OrderHistory{
			OrderId:   e.OrderID,
			EventType: toPbEventType(e.Event),
			CreatedAt: timestamppb.New(e.Timestamp),
		})
	}
	return &pb.OrderHistoryList{
		History: history,
	}
}
