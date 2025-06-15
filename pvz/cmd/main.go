package main

import (
	"context"
	"os"
	"os/signal"
	"pvz-cli/internal/app"
	"pvz-cli/internal/common/observability"
	"pvz-cli/internal/grpc/interceptors"
	"sync"
	"syscall"
)

func main() {
	logger := interceptors.GetLogger().Sugar()
	defer func() {
		if err := interceptors.CloseLogFile(); err != nil {
			logger.Errorf("failed to close log file: %v", err)
		}
	}()

	application := app.New()
	defer application.Shutdown()

	if err := observability.InitTracing(application.Ctx); err != nil {
		logger.Errorf("failed to init tracing: %v", err)
		return
	}
	defer func() {
		if err := observability.ShutdownTracing(context.Background()); err != nil {
			logger.Errorf("failed to shutdown tracing: %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		application.Shutdown()
	}()

	var wg sync.WaitGroup
	wg.Add(4)

	go app.StartGRPCServer(application, &wg)
	go app.StartHTTPGateway(application, &wg)
	go app.StartSwaggerUI(application, &wg)
	go app.StartCLI(application, &wg)

	<-application.Ctx.Done()
	logger.Infow("Shutdown signal received, waiting for all servers…")
	wg.Wait()
	logger.Info("All services stopped, exiting.")
}
