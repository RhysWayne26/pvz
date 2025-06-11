package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

// Application holds shared context and the DI container.
// It does NOT know anything about servers or CLI.
type Application struct {
	Ctx       context.Context
	Cancel    context.CancelFunc
	Container *Container
}

// New wires up the cancellation context and container.
func New() *Application {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return &Application{
		Ctx:       ctx,
		Cancel:    cancel,
		Container: NewContainer(),
	}
}

// Shutdown triggers cancellation; cleanup hooks live in main().
func (a *Application) Shutdown() {
	log.Println("Shutdown signal received")
	a.Cancel()
}
