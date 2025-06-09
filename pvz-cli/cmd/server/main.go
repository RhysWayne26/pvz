package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"pvz-cli/internal/common/config"
	"pvz-cli/internal/common/shutdown"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/grpc/gateway"
	"pvz-cli/internal/grpc/interceptor"
	"pvz-cli/internal/grpc/mappers"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nShutdown initiated")
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
	facadeMapper := mappers.NewDefaultGRPCFacadeMapper()
	grpcRouter := gateway.NewGRPCRouter(facadeMapper, facadeHandler)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.RateLimitInterceptor()),
	)
	pb.RegisterOrdersServiceServer(grpcServer, grpcRouter)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Println("gRPC listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC Serve: %v", err)
		}
	}()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(gateway.GRPCGatewayErrorHandler),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterOrdersServiceHandlerFromEndpoint(ctx, mux, ":50051", opts); err != nil {
		log.Fatalf("failed to register HTTP gateway: %v", err)
	}
	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		log.Println("HTTP gateway listening on :8080")
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP ListenAndServe: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down servers…")
	grpcServer.GracefulStop()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := httpSrv.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP Shutdown error: %v", err)
	}

	log.Println("servers stopped")
}
