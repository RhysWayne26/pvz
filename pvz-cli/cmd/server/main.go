package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/common/observability"
	"pvz-cli/internal/common/shutdown"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/grpc/gateway"
	"pvz-cli/internal/grpc/interceptors"
	"pvz-cli/internal/grpc/mappers"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handleSignals(cancel)

	cfg := config.Load()
	if err := observability.InitTracing(ctx); err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}

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
	facadeMapper := mappers.NewDefaultGRPCFacadeMapper()
	grpcRouter := gateway.NewGRPCRouter(facadeMapper, facadeHandler)

	go func() {
		if err := gateway.RunGRPCServer(
			":50051",
			grpcRouter,
			grpc.ChainUnaryInterceptor(
				interceptors.RecoveryInterceptor(),
				interceptors.CorrelationIDInterceptor(),
				interceptors.TracingInterceptor(),
				interceptors.RateLimitInterceptor(),
				interceptors.LoggingInterceptor(),
			),
		); err != nil {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	go serveSwagger()
	go serveHTTPGateway()
	<-ctx.Done()
	log.Println("shutting down...")
}

func serveSwagger() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("docs/swagger")))
	log.Println("Swagger UI available at http://localhost:8082/")
	if err := http.ListenAndServe(":8082", mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("swagger UI server error: %v", err)
	}
}

func serveHTTPGateway() {
	if err := gateway.RunHTTPGateway(":50051", ":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP gateway error: %v", err)
	}
}

func handleSignals(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nShutdown initiated")
		shutdown.Signal()
		cancel()
	}()
}
