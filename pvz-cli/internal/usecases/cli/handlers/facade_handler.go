package handlers

import (
	"pvz-cli/internal/usecases/dto"
)

// FacadeHandler is an interface that provides centralized coordination of all CLI command handlers.
type FacadeHandler interface {
	HandleHelp() error
	HandleAcceptOrder(params dto.AcceptOrderParams, silent bool) error
	HandleReturnOrder(params dto.ReturnOrderParams) error
	HandleProcessOrders(params dto.ProcessOrdersParams) error
	HandleListOrders(params dto.ListOrdersParams) error
	HandleListReturns(params dto.ListReturnsParams) error
	HandleOrderHistory() error
	HandleImportOrders(params dto.ImportOrdersParams) error
	HandleScrollOrders(params dto.ScrollOrdersParams) error
}
