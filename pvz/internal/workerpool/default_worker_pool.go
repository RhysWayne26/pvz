package workerpool

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var _ WorkerPool = (*DefaultWorkerPool)(nil)

// DefaultWorkerPool is an implementation of the WorkerPool interface providing a pool of worker goroutines for task execution.
type DefaultWorkerPool struct {
	tasks        chan func()
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	workerCount  int32
	mutex        sync.Mutex
	stopWorkers  chan struct{}
	shutdownOnce sync.Once
	logger       *slog.Logger
	logFile      *os.File

	totalTasks  int64
	activeTasks int64
	failedTasks int64
	isShutdown  atomic.Bool
}

// NewDefaultWorkerPool initializes a DefaultWorkerPool with the specified context and options; sets up workers, a task queue, logging, and statistics tracking.
func NewDefaultWorkerPool(parentCtx context.Context, opts ...Option) *DefaultWorkerPool {
	cfg := &Config{}
	cfg.applyDefaults()
	for _, opt := range opts {
		opt(cfg)
	}
	logger, f := openLogFile(cfg)
	ctx, cancel := context.WithCancel(parentCtx)
	capacity := cfg.WorkerCount * cfg.QueueFactor
	p := &DefaultWorkerPool{
		tasks:       make(chan func(), capacity),
		ctx:         ctx,
		cancel:      cancel,
		workerCount: int32(cfg.WorkerCount),
		stopWorkers: make(chan struct{}, cfg.WorkerCount),
		logger:      logger,
		logFile:     f,
	}
	for i := 0; i < cfg.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	go p.startStatsLogger(cfg.StatsInterval)
	return p
}

// Submit adds a task to the worker pool for execution, rejecting it if the pool is shut down or the context is canceled.
func (p *DefaultWorkerPool) Submit(task func()) {
	if p.IsShutdown() {
		p.logger.Warn("task rejected - pool is shutdown")
		return
	}
	atomic.AddInt64(&p.totalTasks, 1)
	p.logger.Debug("task queued", "queue_size", len(p.tasks))
	select {
	case p.tasks <- task:
	case <-p.ctx.Done():
		p.logger.Warn("task rejected during shutdown")
	}
}

// SubmitWithContext queues a task for execution with a context, returning an error if the context or pool is closed.
func (p *DefaultWorkerPool) SubmitWithContext(ctx context.Context, task func()) error {
	if p.IsShutdown() {
		return ErrPoolShuttingDown
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return ErrPoolShuttingDown
	case p.tasks <- task:
		atomic.AddInt64(&p.totalTasks, 1)
		p.logger.Debug("task queued with context",
			"total_tasks", atomic.LoadInt64(&p.totalTasks),
		)
		return nil
	}
}

// Shutdown gracefully shuts down the worker pool, ensuring all tasks are completed and resources are released.
func (p *DefaultWorkerPool) Shutdown() {
	p.shutdownOnce.Do(func() {
		p.logger.Info("initiating worker pool shutdown")
		p.isShutdown.Store(true)
		p.cancel()
		close(p.tasks)
		p.wg.Wait()
		p.logger.Info("worker pool shutdown completed")
		if err := p.logFile.Close(); err != nil {
			p.logger.Error("failed to close pool log file", "error", err)
		}
	})
}

// ShutdownWithTimeout gracefully shuts down the worker pool within a specified timeout, returning an error on timeout.
func (p *DefaultWorkerPool) ShutdownWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		p.Shutdown()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return ErrShutdownTimeout
	}
}

// SetWorkerCount adjusts the number of workers in the pool, scaling up or down based on the specified target count.
func (p *DefaultWorkerPool) SetWorkerCount(count int) {
	if count < 0 {
		p.logger.Warn("invalid worker count", "count", count)
		return
	}
	if p.IsShutdown() {
		p.logger.Warn("cannot set worker count - pool is shutdown")
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	cur := int(atomic.LoadInt32(&p.workerCount))
	p.logger.Info("adjusting worker count",
		"current", cur,
		"target", count,
	)
	switch {
	case count > cur:
		for i := cur; i < count; i++ {
			p.wg.Add(1)
			go p.worker(i)
		}
		p.logger.Info("workers scaled up",
			"added", count-cur,
			"total", count,
		)
	case count < cur:
		toStop := cur - count
		for i := 0; i < toStop; i++ {
			p.stopWorkers <- struct{}{}
		}
		p.logger.Info("workers scaling down",
			"removing", toStop,
			"target", count,
		)
	}

	atomic.StoreInt32(&p.workerCount, int32(count))
}

// GetStats retrieves and returns current statistics of the worker pool including worker count, queue size, and task metrics.
func (p *DefaultWorkerPool) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"worker_count": atomic.LoadInt32(&p.workerCount),
		"queue_size":   len(p.tasks),
		"active_tasks": atomic.LoadInt64(&p.activeTasks),
		"total_tasks":  atomic.LoadInt64(&p.totalTasks),
		"failed_tasks": atomic.LoadInt64(&p.failedTasks),
		"is_shutdown":  p.IsShutdown(),
	}
	p.logger.Debug("pool stats requested", "stats", stats)
	return stats
}

// IsShutdown returns true if the worker pool has been shut down, indicating no further tasks can be submitted.
func (p *DefaultWorkerPool) IsShutdown() bool {
	return p.isShutdown.Load()
}

// worker is a goroutine function responsible for executing tasks from the task queue until shutdown signals are received.
func (p *DefaultWorkerPool) worker(id int) {
	defer func() {
		p.logger.Info("worker exiting", "id", id)
		p.wg.Done()
	}()
	p.logger.Info("worker started", "id", id)
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			p.logger.Debug("worker picked up task",
				"id", id,
				"active_tasks", atomic.AddInt64(&p.activeTasks, 1),
				"queue_size", len(p.tasks),
			)
			success := p.execTask(id, task)
			atomic.AddInt64(&p.activeTasks, -1)
			if !success {
				p.logger.Warn("task failed", "worker", id)
			}
		case <-p.stopWorkers:
			p.logger.Info("worker stop signal", "id", id)
			return
		case <-p.ctx.Done():
			p.logger.Info("worker context canceled", "id", id)
			return
		}
	}
}

func (p *DefaultWorkerPool) execTask(id int, task func()) bool {
	start := time.Now()
	success := true
	defer func() {
		if r := recover(); r != nil {
			success = false
			atomic.AddInt64(&p.failedTasks, 1)
			p.logger.Error("panic in task",
				"worker", id,
				"panic", r,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	}()
	task()
	p.logger.Debug("worker finished task",
		"worker", id,
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return success
}

func (p *DefaultWorkerPool) startStatsLogger(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.logger.Info("worker pool stats",
					"active_workers", atomic.LoadInt32(&p.workerCount),
					"queue_size", len(p.tasks),
					"active_tasks", atomic.LoadInt64(&p.activeTasks),
					"total_tasks", atomic.LoadInt64(&p.totalTasks),
					"failed_tasks", atomic.LoadInt64(&p.failedTasks),
				)
			case <-p.ctx.Done():
				p.logger.Info("stats logger shutting down")
				return
			}
		}
	}()
}
