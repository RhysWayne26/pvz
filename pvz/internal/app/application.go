package app

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os/signal"
	"pvz-cli/docs/swagger"
	"pvz-cli/internal/cli"
	climappers "pvz-cli/internal/cli/mappers"
	"pvz-cli/internal/grpc/gateway"
	"pvz-cli/internal/grpc/interceptors"
	grpcmappers "pvz-cli/internal/grpc/mappers"
	"pvz-cli/internal/workerpool"
	"sync"
	"syscall"
	"time"
)

// The Application holds shared context and the DI container.
type Application struct {
	Ctx       context.Context
	Cancel    context.CancelFunc
	Container *Container
	Pool      workerpool.WorkerPool
}

// New wires up the cancellation context and container.
func New() *Application {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	pool := workerpool.NewDefaultWorkerPool(
		ctx,
		workerpool.WithWorkerCount(2),
		workerpool.WithStatsInterval(10*time.Second),
		workerpool.WithQueueFactor(3),
	)

	return &Application{
		Ctx:       ctx,
		Cancel:    cancel,
		Container: NewContainer(pool),
		Pool:      pool,
	}
}

// Shutdown triggers cancellation; cleanup hooks live in main().
func (a *Application) Shutdown() {
	log.Println("Shutdown signal received")
	a.Cancel()
	log.Println("Shutting down worker pool")
	a.Pool.Shutdown()
	log.Println("Worker pool is shutdown")
}

// StartGRPCServer launches the main gRPC server on :50051 with interceptors for logging, tracing, validation, etc.
func StartGRPCServer(app *Application, port string, wg *sync.WaitGroup) {
	defer wg.Done()
	mapper := grpcmappers.NewDefaultGRPCFacadeMapper()
	router := gateway.NewGRPCRouter(mapper, app.Container.FacadeHandler)

	err := gateway.RunGRPCServer(
		app.Ctx,
		port,
		router,
		grpc.ChainUnaryInterceptor(
			interceptors.ValidationInterceptor(),
			interceptors.RecoveryInterceptor(),
			interceptors.CorrelationIDInterceptor(),
			interceptors.TracingInterceptor(),
			interceptors.RateLimitInterceptor(),
			interceptors.LoggingInterceptor(),
		),
	)
	if err != nil && app.Ctx.Err() == nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

// StartHTTPGateway starts the gRPC-Gateway proxy on :8080, routing HTTP requests to the gRPC backend.
func StartHTTPGateway(app *Application, wg *sync.WaitGroup) {
	defer wg.Done()
	err := gateway.RunHTTPGateway(app.Ctx, ":50051", ":50052", ":8080")
	if err != nil && app.Ctx.Err() == nil {
		log.Fatalf("HTTP gateway error: %v", err)
	}
}

// StartSwaggerUI serves the embedded Swagger UI from / on :8082.
func StartSwaggerUI(app *Application, wg *sync.WaitGroup) {
	defer wg.Done()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(swagger.FS)))

	srv := &http.Server{
		Addr:              ":8082",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-app.Ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Swagger UI shutdown error: %v", err)
		}
	}()
	log.Printf("Swagger UI started at http://localhost:8082/")
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Swagger UI error: %v", err)
	}
}

// StartCLI runs the interactive CLI interface using the shared DI container.
func StartCLI(app *Application, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("CLI started")
	mapper := climappers.NewDefaultFacadeMapper()
	router := cli.NewRouter(app.Container.FacadeHandler, mapper)
	router.Run(app.Ctx, app.Shutdown)
	log.Println("CLI finished")
}

func StartAdminGRPCServer(app *Application, port string, wg *sync.WaitGroup) {
	defer wg.Done()
	router := gateway.NewAdminGRPCRouter(app.Pool)
	err := gateway.RunAdminGRPCServer(
		app.Ctx,
		port,
		router,
		grpc.ChainUnaryInterceptor(
			interceptors.ValidationInterceptor(),
			interceptors.RecoveryInterceptor(),
		),
	)
	if err != nil && app.Ctx.Err() == nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
