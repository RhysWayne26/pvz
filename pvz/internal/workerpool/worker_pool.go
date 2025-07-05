package workerpool

import (
	"context"
	"time"
)

type WorkerPool interface {
	Submit(task func())
	SubmitWithContext(ctx context.Context, task func()) error
	Shutdown()
	ShutdownWithTimeout(timeout time.Duration) error
	SetWorkerCount(count int)
	GetStats() map[string]interface{}
	IsShutdown() bool
}
