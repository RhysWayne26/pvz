package main

import (
	"fmt"
	"os"
	"os/signal"
	"pvz-cli/internal/usecases/cli/handlers"
	"pvz-cli/internal/usecases/strategies"
	"syscall"

	"pvz-cli/cmd/cli"
	"pvz-cli/internal/config"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/shutdown"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/validators"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nShutdown initiated, other saves will not start.")
		shutdown.Signal()
	}()

	cfg := config.Load()
	store := storage.NewJSONStorage(cfg.Path)

	orderRepo := repositories.NewSnapshotOrderRepository(store)
	returnRepo := repositories.NewSnapshotReturnRepository(store)
	historyRepo := repositories.NewSnapshotHistoryRepository(store)

	orderValidator := validators.NewDefaultOrderValidator()
	packageValidator := validators.NewDefaultPackageValidator()

	pricingStrategy := strategies.NewDefaultPricingStrategy()

	historySvc := services.NewDefaultHistoryService(historyRepo)
	packagePricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	orderSvc := services.NewDefaultOrderService(orderRepo, returnRepo, packagePricingSvc, historySvc, orderValidator)

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)

	router := cli.NewRouter(facadeHandler)
	router.Run()
}
