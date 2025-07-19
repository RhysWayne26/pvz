package handlers

import (
	"pvz-cli/internal/usecases/services"
	"pvz-cli/pkg/cache"
	"pvz-cli/pkg/metrics"
)

var _ FacadeHandler = (*DefaultFacadeHandler)(nil)

// DefaultFacadeHandler is the default implementation of the FacadeHandler interface.
type DefaultFacadeHandler struct {
	orderService   services.OrderService
	historyService services.HistoryService
	responsesCache cache.Cache[string, any]
	metrics        metrics.HandlerMetrics
}

// NewDefaultFacadeHandler constructs a new DefaultFacadeHandler with the provided services.
func NewDefaultFacadeHandler(
	orderSvc services.OrderService,
	historySvc services.HistoryService,
	responsesCache cache.Cache[string, any],
	metrics metrics.HandlerMetrics,
) *DefaultFacadeHandler {
	return &DefaultFacadeHandler{
		orderService:   orderSvc,
		historyService: historySvc,
		responsesCache: responsesCache,
		metrics:        metrics,
	}
}
