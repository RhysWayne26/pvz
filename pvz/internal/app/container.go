package app

import (
	"log/slog"
	"os"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
	"pvz-cli/internal/workerpool"
	"pvz-cli/pkg/clock"
	"time"
)

// Container holds all shared business-level dependencies: configuration, repositories, services, and the facade handler.
type Container struct {
	config         *config.Config
	orderService   services.OrderService
	historyService services.HistoryService
	facadeHandler  handlers.FacadeHandler
}

// NewContainer returns a new instance of an application container
func NewContainer(pool workerpool.WorkerPool) *Container {
	cfg := config.Load()
	var orderRepo repositories.OrderRepository
	var historyRepo repositories.HistoryRepository

	switch {
	case cfg.DB != nil && cfg.DB.WriteDSN != "":
		client, err := db.NewDefaultPGXClient(cfg.DB.ReadDSN, cfg.DB.WriteDSN)
		if err != nil {
			slog.Error("failed to init DB client", "error", err)
			os.Exit(1)
		}
		client.SetConnectionSettings(20, 10, time.Hour, 30*time.Minute)
		orderRepo = repositories.NewPGOrderRepository(client)
		historyRepo = repositories.NewPGHistoryRepository(client)

	case cfg.File != nil && cfg.File.Path != "":
		fileStorage := storage.NewJSONStorage(cfg.File.Path)
		orderRepo = repositories.NewSnapshotOrderRepository(fileStorage)
		historyRepo = repositories.NewSnapshotHistoryRepository(fileStorage)

	default:
		slog.Error("No valid storage config provided")
		os.Exit(1)
	}

	clk := &clock.RealClock{}

	orderValidator := validators.NewDefaultOrderValidator(clk)
	packageValidator := validators.NewDefaultPackageValidator()
	pricingStrategy := strategies.NewDefaultPricingStrategy()

	historySvc := services.NewDefaultHistoryService(historyRepo)
	pricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	orderSvc := services.NewDefaultOrderService(clk, pool, orderRepo, pricingSvc, historySvc, orderValidator)

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)

	return &Container{
		config:         cfg,
		orderService:   orderSvc,
		historyService: historySvc,
		facadeHandler:  facadeHandler,
	}
}
