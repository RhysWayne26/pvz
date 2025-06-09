package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbProcessOrdersRequest maps a gRPC ProcessOrdersRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbProcessOrdersRequest(in *pb.ProcessOrdersRequest) requests.ProcessOrdersRequest {
	var action requests.ProcessAction
	switch in.Action {
	case pb.ActionType_ACTION_TYPE_ISSUE:
		action = requests.ActionIssue
	case pb.ActionType_ACTION_TYPE_RETURN:
		action = requests.ActionReturn
	default:
		action = "unknown"
	}
	return requests.ProcessOrdersRequest{
		UserID:   in.UserId,
		OrderIDs: in.OrderIds,
		Action:   action,
	}
}

// ToPbProcessResult maps the internal ProcessOrdersResponse to a gRPC ProcessResult.
func (f *DefaultGRPCFacadeMapper) ToPbProcessResult(res responses.ProcessOrdersResponse) *pb.ProcessResult {
	var failedIDs []uint64
	for _, report := range res.Failed {
		failedIDs = append(failedIDs, report.OrderID)
	}
	return &pb.ProcessResult{
		Processed: res.Processed,
		Errors:    failedIDs,
	}
}
