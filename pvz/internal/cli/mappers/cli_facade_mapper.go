package mappers

import (
	"pvz-cli/internal/cli/params"
	"pvz-cli/internal/usecases/requests"
)

// CLIFacadeMapper defines an interface for mapping CLI input parameters into internal request models used by business logic.
type CLIFacadeMapper interface {
	// MapAcceptOrderParams maps accept-order CLI parameters to an internal request.
	MapAcceptOrderParams(params.AcceptOrderParams) (requests.AcceptOrderRequest, error)

	// MapScrollOrdersParams maps scroll-orders CLI parameters to a scroll request.
	MapScrollOrdersParams(params.ScrollOrdersParams) (requests.OrdersFilterRequest, error)

	// MapListOrdersParams maps list-orders CLI parameters to a filtering request.
	MapListOrdersParams(params.ListOrdersParams) (requests.OrdersFilterRequest, error)

	// MapProcessOrdersParams maps process-orders CLI parameters to a process request.
	MapProcessOrdersParams(params.ProcessOrdersParams) (requests.ProcessOrdersRequest, error)

	// MapImportOrdersParams maps import-orders CLI parameters to a batch request.
	MapImportOrdersParams(params.ImportOrdersParams) (requests.ImportOrdersRequest, error)

	// MapListReturnsParams maps list-returns CLI parameters to a returns request.
	MapListReturnsParams(params.ListReturnsParams) (requests.OrdersFilterRequest, error)

	// MapReturnOrderParams maps return-order CLI parameters to a return request.
	MapReturnOrderParams(params.ReturnOrderParams) (requests.ReturnOrderRequest, error)

	// MapOrderHistoryParams maps list-orders CLI parameters to a filtering request.
	MapOrderHistoryParams(params.OrderHistoryParams) (requests.OrderHistoryRequest, error)
}
