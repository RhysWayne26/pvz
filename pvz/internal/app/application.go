package app

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pvz-cli/docs/swagger"
	"pvz-cli/internal/cli"
	climappers "pvz-cli/internal/cli/mappers"
	"pvz-cli/internal/common/observability"
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
	ctx       context.Context
	cancel    context.CancelFunc
	container *Container
	pool      workerpool.WorkerPool
	logger    *zap.SugaredLogger
	wg        sync.WaitGroup
}

// New wires up the cancellation context and container.
func New(logger *zap.SugaredLogger) *Application {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	_ = godotenv.Load()
	if err := observability.InitTracing(ctx); err != nil {
		logger.Errorf("failed to init tracing: %v", err)
	}

	pool := workerpool.NewDefaultWorkerPool(
		ctx,
		workerpool.WithWorkerCount(2),
		workerpool.WithStatsInterval(10*time.Second),
		workerpool.WithQueueFactor(3),
	)

	return &Application{
		ctx:       ctx,
		cancel:    cancel,
		container: NewContainer(pool),
		pool:      pool,
		logger:    logger,
		wg:        sync.WaitGroup{},
	}
}

func (a *Application) Run() {
	grpcPort := os.Getenv("GRPC_PORT")
	adminGRPCPort := os.Getenv("ADMIN_GRPC_PORT")
	services := []func(){
		func() { a.StartGRPCServer(grpcPort) },
		func() { a.StartHTTPGateway() },
		func() { a.StartSwaggerUI() },
		func() { a.StartCLI() },
		func() { a.StartAdminGRPCServer(adminGRPCPort) },
		func() { a.StartMetricsServer() },
	}
	if a.container.outboxDispatcher != nil {
		services = append(services, func() {
			a.StartOutboxDispatcher()
		})
	}

	a.wg.Add(len(services))
	for _, service := range services {
		go service()
	}
	a.Wait()
}

func (a *Application) Wait() {
	<-a.ctx.Done()
	a.logger.Info("Shutdown signal received, waiting for all servers…")
	a.wg.Wait()
	a.logger.Info("All services stopped, exiting.")
}

// Shutdown triggers cancellation; cleanup hooks live in main().
func (a *Application) Shutdown() {
	log.Println("Shutdown signal received")
	a.cancel()
	a.container.shutdownOutbox()
	defer func() {
		if err := observability.ShutdownTracing(context.Background()); err != nil {
			a.logger.Errorf("failed to shutdown tracing: %v", err)
		}
	}()
	log.Println("Shutting down worker pool")
	a.pool.Shutdown()
	log.Println("Worker pool is shutdown")
}

// StartGRPCServer launches the main gRPC server on :50051 with interceptors for logging, tracing, validation, etc.
func (a *Application) StartGRPCServer(port string) {
	defer a.wg.Done()
	mapper := grpcmappers.NewDefaultGRPCFacadeMapper()
	router := gateway.NewGRPCRouter(mapper, a.container.facadeHandler)

	err := gateway.RunGRPCServer(
		a.ctx,
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
	if err != nil && a.ctx.Err() == nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

// StartHTTPGateway starts the gRPC-Gateway proxy on :8080, routing HTTP requests to the gRPC backend.
func (a *Application) StartHTTPGateway() {
	defer a.wg.Done()
	err := gateway.RunHTTPGateway(a.ctx, ":50051", ":50052", ":8080")
	if err != nil && a.ctx.Err() == nil {
		log.Fatalf("HTTP gateway error: %v", err)
	}
}

// StartSwaggerUI serves the embedded Swagger UI from / on :8082.
func (a *Application) StartSwaggerUI() {
	defer a.wg.Done()
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
		<-a.ctx.Done()
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
func (a *Application) StartCLI() {
	defer a.wg.Done()
	log.Println("CLI started")
	mapper := climappers.NewDefaultFacadeMapper()
	router := cli.NewRouter(a.container.facadeHandler, mapper)
	router.Run(a.ctx, a.Shutdown)
	log.Println("CLI finished")
}

// StartAdminGRPCServer starts the admin gRPC server on the specified port with validation and recovery interceptors.
func (a *Application) StartAdminGRPCServer(port string) {
	defer a.wg.Done()
	router := gateway.NewAdminGRPCRouter(a.pool)
	err := gateway.RunAdminGRPCServer(
		a.ctx,
		port,
		router,
		grpc.ChainUnaryInterceptor(
			interceptors.ValidationInterceptor(),
			interceptors.RecoveryInterceptor(),
		),
	)
	if err != nil && a.ctx.Err() == nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

func (a *Application) StartOutboxDispatcher() {
	defer a.wg.Done()
	if err := a.container.outboxDispatcher.Dispatch(a.ctx); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Errorf("outbox dispatcher stopped: %v", err)
	}
}

func (a *Application) StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("metrics server error: %v", err)
		}
	}()
}

func (a *Application) AddToWaitGroup(n int) {
	a.wg.Add(n)
}
