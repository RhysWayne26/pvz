package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// GRPCFacadeMapper defines conversion between protobuf messages and internal request/response DTOs.
type GRPCFacadeMapper interface {
	// FromPbAcceptOrderRequest maps protobuf AcceptOrderRequest to internal AcceptOrderRequest.
	FromPbAcceptOrderRequest(*pb.AcceptOrderRequest) (requests.AcceptOrderRequest, error)

	// FromPbReturnOrderRequest maps protobuf OrderIdRequest to internal ReturnOrderRequest.
	FromPbReturnOrderRequest(*pb.OrderIdRequest) requests.ReturnOrderRequest

	// FromPbProcessOrdersRequest maps protobuf ProcessOrdersRequest to internal ProcessOrdersRequest.
	FromPbProcessOrdersRequest(*pb.ProcessOrdersRequest) requests.ProcessOrdersRequest

	// FromPbListOrdersRequest maps protobuf OrdersFilterRequest to internal OrdersFilterRequest.
	FromPbListOrdersRequest(*pb.ListOrdersRequest) requests.OrdersFilterRequest

	// FromPbListReturnsRequest maps protobuf ListReturnsRequest to internal ListReturnsRequest.
	FromPbListReturnsRequest(*pb.ListReturnsRequest) requests.OrdersFilterRequest

	// FromPbImportOrdersRequest maps protobuf ImportOrdersRequest to internal ImportOrdersRequest.
	FromPbImportOrdersRequest(*pb.ImportOrdersRequest) requests.ImportOrdersRequest

	// ToPbAcceptOrderResponse maps internal AcceptOrderResponse to protobuf OrderResponse.
	ToPbAcceptOrderResponse(res responses.AcceptOrderResponse) *pb.OrderResponse

	// ToPbReturnOrderResponse maps internal ReturnOrderResponse to protobuf OrderResponse.
	ToPbReturnOrderResponse(res responses.ReturnOrderResponse) *pb.OrderResponse

	// ToPbProcessResult maps internal ProcessOrdersResponse to protobuf ProcessResult.
	ToPbProcessResult(res responses.ProcessOrdersResponse) *pb.ProcessResult

	// ToPbOrdersList maps internal ListOrdersResponse to protobuf OrdersList.
	ToPbOrdersList(res responses.ListOrdersResponse) *pb.OrdersList

	// ToPbReturnsList maps internal ListReturnsResponse to protobuf ReturnsList.
	ToPbReturnsList(res responses.ListOrdersResponse) *pb.ReturnsList

	// ToPbOrderHistoryList maps internal OrderHistoryResponse to protobuf OrderHistoryList.
	ToPbOrderHistoryList(res responses.OrderHistoryResponse) *pb.OrderHistoryList

	// ToPbImportResult maps internal ImportOrdersResponse to protobuf ImportResult.
	ToPbImportResult(res responses.ImportOrdersResponse) *pb.ImportResult
}
