package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	"pvz-cli/internal/app"
	"pvz-cli/internal/common/observability"
	"pvz-cli/internal/grpc/gateway"
	"pvz-cli/internal/grpc/interceptors"
	"pvz-cli/internal/grpc/mappers"
)

func main() {
	application := app.New()
	defer application.Shutdown()
	if err := observability.InitTracing(application.Ctx); err != nil {
		log.Printf("failed to init tracing: %v", err)
		application.Shutdown()
	}
	var wg sync.WaitGroup
	wg.Add(3)
	grpcMapper := mappers.NewDefaultGRPCFacadeMapper()
	grpcRouter := gateway.NewGRPCRouter(grpcMapper, application.Container.FacadeHandler)
	go func() {
		defer wg.Done()
		if err := gateway.RunGRPCServer(
			application.Ctx,
			":50051",
			grpcRouter,
			grpc.ChainUnaryInterceptor(
				interceptors.RecoveryInterceptor(),
				interceptors.CorrelationIDInterceptor(),
				interceptors.TracingInterceptor(),
				interceptors.RateLimitInterceptor(),
				interceptors.LoggingInterceptor(),
			),
		); err != nil && application.Ctx.Err() == nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := gateway.RunHTTPGateway(
			application.Ctx,
			":50051",
			":8080",
		); err != nil && application.Ctx.Err() == nil {
			log.Fatalf("HTTP gateway error: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(http.Dir("docs/swagger")))
		srv := &http.Server{
			Addr:              ":8082",
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		}
		go func() {
			<-application.Ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Printf("Swagger UI shutdown error: %v", err)
			}
		}()
		log.Println("Swagger UI available at http://localhost:8082/")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Swagger UI error: %v", err)
		}
	}()
	<-application.Ctx.Done()
	log.Println("Shutdown signal received, waiting for all servers…")
	wg.Wait()
	log.Println("All services stopped, exiting.")
	os.Exit(0)
}
