package main

import (
	"pvz-cli/internal/app"
	"pvz-cli/internal/grpc/interceptors"
)

func main() {
	logger := interceptors.GetLogger().Sugar()
	defer func() {
		if err := interceptors.CloseLogFile(); err != nil {
			logger.Errorf("failed to close log file: %v", err)
		}
	}()
	application := app.New(logger)
	defer application.Shutdown()

	application.Run()
	application.Wait()
}
