package mappers

import (
	"pvz-cli/internal/common/apperrors"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FromPbProcessOrdersRequest maps a gRPC ProcessOrdersRequest to the internal request model.
func (f *DefaultGRPCFacadeMapper) FromPbProcessOrdersRequest(in *pb.ProcessOrdersRequest) (requests.ProcessOrdersRequest, error) {
	if err := providedUserIDCheck(in.UserId); err != nil {
		return requests.ProcessOrdersRequest{}, err
	}
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
	}, nil
}

// ToPbProcessResult maps the internal ProcessOrdersResponse to a gRPC ProcessResult.
func (f *DefaultGRPCFacadeMapper) ToPbProcessResult(res responses.ProcessOrdersResponse) *pb.ProcessResult {
	var failedOrders []*pb.FailedBatchedOrder
	for _, report := range res.Failed {
		failedOrders = append(failedOrders, &pb.FailedBatchedOrder{
			OrderId: report.OrderID,
			Code:    apperrors.CodeFromError(report.Error),
			Reason:  apperrors.MessageFromError(report.Error),
		})
	}
	return &pb.ProcessResult{
		Processed: res.Processed,
		Errors:    failedOrders,
	}
}
