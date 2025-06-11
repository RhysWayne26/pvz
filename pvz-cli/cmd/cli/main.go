package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"pvz-cli/internal/app"
	"pvz-cli/internal/cli"
	"pvz-cli/internal/cli/mappers"
)

func main() {
	application := app.New()
	var once sync.Once
	shutdown := func() {
		once.Do(func() {
			log.Println("Shutdown initiated (CLI).")
			application.Shutdown()
		})
	}
	defer shutdown()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		shutdown()
	}()
	cliMapper := mappers.NewDefaultFacadeMapper()
	router := cli.NewRouter(
		application.Container.FacadeHandler,
		cliMapper,
	)
	router.Run(application.Ctx)
	log.Println("CLI execution finished, exiting.")
}
