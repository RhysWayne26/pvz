package handlers

import (
	"context"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/responses"
)

// FacadeHandler is an interface that provides centralized coordination of all CLI command handlers.
type FacadeHandler interface {
	HandleAcceptOrder(ctx context.Context, req requests.AcceptOrderRequest, silent bool) (responses.AcceptOrderResponse, error)
	HandleReturnOrder(ctx context.Context, req requests.ReturnOrderRequest) (responses.ReturnOrderResponse, error)
	HandleProcessOrders(ctx context.Context, req requests.ProcessOrdersRequest) (responses.ProcessOrdersResponse, error)
	HandleListOrders(ctx context.Context, req requests.ListOrdersRequest) (responses.ListOrdersResponse, error)
	HandleListReturns(ctx context.Context, req requests.ListReturnsRequest) (responses.ListReturnsResponse, error)
	HandleOrderHistory(ctx context.Context) (responses.OrderHistoryResponse, error)
	HandleImportOrders(ctx context.Context, req requests.ImportOrdersRequest) (responses.ImportOrdersResponse, error)

	// HandleScrollOrders and HandleHelp are CLI-only, intentionally without context
	HandleScrollOrders(req requests.ScrollOrdersRequest) error
	HandleHelp() error
}
