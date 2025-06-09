package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"pvz-cli/internal/cli"
	"pvz-cli/internal/cli/mappers"
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/common/shutdown"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nShutdown initiated, other saves will not start.")
		shutdown.Signal()
		cancel()
	}()

	cfg := config.Load()
	store := storage.NewJSONStorage(cfg.Path)

	orderRepo := repositories.NewSnapshotOrderRepository(store)
	historyRepo := repositories.NewSnapshotHistoryRepository(store)

	orderValidator := validators.NewDefaultOrderValidator()
	packageValidator := validators.NewDefaultPackageValidator()

	pricingStrategy := strategies.NewDefaultPricingStrategy()

	historySvc := services.NewDefaultHistoryService(historyRepo)
	packagePricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	orderSvc := services.NewDefaultOrderService(orderRepo, packagePricingSvc, historySvc, orderValidator)

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)
	facadeMapper := mappers.NewDefaultFacadeMapper()

	router := cli.NewRouter(facadeHandler, facadeMapper)
	router.Run(ctx)
}
