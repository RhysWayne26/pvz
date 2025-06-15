package handlers

import (
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
