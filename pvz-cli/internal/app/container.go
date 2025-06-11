package app

import (
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
)

// Container holds all shared business-level dependencies: configuration, repositories, services, and the facade handler.
type Container struct {
	Config         *config.Config
	OrderService   services.OrderService
	HistoryService services.HistoryService
	FacadeHandler  handlers.FacadeHandler
}

// NewContainer returns a new instance of an application container
func NewContainer() *Container {
	cfg := config.Load()

	store := storage.NewJSONStorage(cfg.Path)
	orderRepo := repositories.NewSnapshotOrderRepository(store)
	historyRepo := repositories.NewSnapshotHistoryRepository(store)

	orderValidator := validators.NewDefaultOrderValidator()
	packageValidator := validators.NewDefaultPackageValidator()
	pricingStrategy := strategies.NewDefaultPricingStrategy()

	historySvc := services.NewDefaultHistoryService(historyRepo)
	pricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	orderSvc := services.NewDefaultOrderService(orderRepo, pricingSvc, historySvc, orderValidator)

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)

	return &Container{
		Config:         cfg,
		OrderService:   orderSvc,
		HistoryService: historySvc,
		FacadeHandler:  facadeHandler,
	}
}
