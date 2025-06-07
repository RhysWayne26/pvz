package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/responses"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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
