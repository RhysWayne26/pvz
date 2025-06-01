package handlers

import (
	"pvz-cli/internal/usecases/dto"
	"pvz-cli/internal/usecases/services"
)

// DefaultFacadeHandler is the default implementation of the FacadeHandler interface.
type DefaultFacadeHandler struct {
	orderService   services.OrderService
	historyService services.HistoryService
}

// NewDefaultFacadeHandler constructs a new DefaultFacadeHandler with the provided services.
func NewDefaultFacadeHandler(
	orderSvc services.OrderService,
	historySvc services.HistoryService,
) *DefaultFacadeHandler {
	return &DefaultFacadeHandler{
		orderService:   orderSvc,
		historyService: historySvc,
	}
}

// HandleHelp processes the help command and displays usage information.
func (f *DefaultFacadeHandler) HandleHelp() error {
	return NewHelpCommandHandler().Handle()
}

// HandleAcceptOrder processes a single order acceptance command.
func (f *DefaultFacadeHandler) HandleAcceptOrder(params dto.AcceptOrderParams, silent bool) error {
	return NewAcceptOrderHandler(params, f.orderService, silent).Handle()
}

// HandleReturnOrder processes a single order return to courier.
func (f *DefaultFacadeHandler) HandleReturnOrder(params dto.ReturnOrderParams) error {
	return NewReturnOrderHandler(params, f.orderService).Handle()
}

// HandleProcessOrders processes multiple orders for issue or client return.
func (f *DefaultFacadeHandler) HandleProcessOrders(params dto.ProcessOrdersParams) error {
	return NewProcessOrdersHandler(params, f.orderService).Handle()
}

// HandleListOrders lists orders with optional filters and pagination.
func (f *DefaultFacadeHandler) HandleListOrders(params dto.ListOrdersParams) error {
	return NewListOrdersHandler(params, f.orderService).Handle()
}

// HandleListReturns lists client returns.
func (f *DefaultFacadeHandler) HandleListReturns(params dto.ListReturnsParams) error {
	return NewListReturnsHandler(params, f.orderService).Handle()
}

// HandleOrderHistory displays the order event history.
func (f *DefaultFacadeHandler) HandleOrderHistory() error {
	return NewOrderHistoryHandler(f.historyService).Handle()
}

// HandleImportOrders processes import of orders from a JSON file.
func (f *DefaultFacadeHandler) HandleImportOrders(params dto.ImportOrdersParams) error {
	return NewImportOrdersHandler(params, f.orderService).Handle()
}

// HandleScrollOrders performs paginated output for orders (infinite scroll).
func (f *DefaultFacadeHandler) HandleScrollOrders(params dto.ScrollOrdersParams) error {
	return NewScrollOrdersHandler(params, f.orderService).Handle()
}
