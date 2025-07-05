package workerpool

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDefaultWorkerPool_ParallelExecution verifies the DefaultWorkerPool can execute tasks in parallel with expected results.
func TestDefaultWorkerPool_ParallelExecution(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := NewDefaultWorkerPool(
		ctx,
		WithWorkerCount(2),
		WithQueueFactor(1),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	var executed int32
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		p.Submit(func() {
			atomic.AddInt32(&executed, 1)
			wg.Done()
		})
	}
	wg.Wait()
	require.Equal(t, int32(2), executed)
	stats := p.GetStats()
	require.Equal(t, int64(2), stats["total_tasks"].(int64))
}

// TestDefaultWorkerPool_TwoWorkers_SubmitWithContextCancel tests submitting a task with a canceled context to a worker pool.
func TestDefaultWorkerPool_TwoWorkers_SubmitWithContextCancel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := NewDefaultWorkerPool(
		ctx,
		WithWorkerCount(1),
		WithQueueFactor(1),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	var wg sync.WaitGroup
	wg.Add(1)
	p.Submit(func() {
		wg.Wait()
	})
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.SubmitWithContext(cancelCtx, func() {
		t.Error("Task should not execute with canceled context")
	})
	wg.Done()
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

// TestDefaultWorkerPool_TwoWorkers_TaskCount verifies that a default worker pool with two workers executes all submitted tasks.
func TestDefaultWorkerPool_TwoWorkers_TaskCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := NewDefaultWorkerPool(
		ctx, WithWorkerCount(2),
		WithQueueFactor(1),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	var count int32
	total := 10
	wg := sync.WaitGroup{}
	wg.Add(total)
	for i := 0; i < total; i++ {
		p.Submit(func() {
			atomic.AddInt32(&count, 1)
			wg.Done()
		})
	}
	wg.Wait()
	stats := p.GetStats()
	require.Equal(t, int32(total), count)
	require.Equal(t, int64(total), stats["total_tasks"].(int64))
}

// TestDefaultWorkerPool_FIFOTaskOrder verifies that tasks submitted to DefaultWorkerPool are executed in FIFO order.
func TestDefaultWorkerPool_FIFOTaskOrder(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithQueueFactor(10),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	var mu sync.Mutex
	var order []int
	total := 5
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		idx := i
		p.Submit(func() {
			mu.Lock()
			order = append(order, idx)
			mu.Unlock()
			wg.Done()
		})
	}
	wg.Wait()
	expected := make([]int, total)
	for i := 0; i < total; i++ {
		expected[i] = i
	}
	require.Equal(t, expected, order)
}

// TestSubmitAfterShutdown_IsNoop verifies that submitting tasks after shutdown is ignored.
func TestSubmitAfterShutdown_IsNoop(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	p.Shutdown()
	p.Submit(func() { t.Fail() })
	require.True(t, p.IsShutdown())
	stats := p.GetStats()
	require.Equal(t, int64(0), stats["total_tasks"].(int64))
}

// TestSubmitWithContextAfterShutdown verifies that SubmitWithContext returns ErrPoolShuttingDown after shutdown.
func TestSubmitWithContextAfterShutdown(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	p.Shutdown()
	err := p.SubmitWithContext(context.Background(), func() { t.Fail() })
	require.ErrorIs(t, err, ErrPoolShuttingDown)
}

// TestSubmitWithContext_Success validates that a task submitted with context executes successfully in the worker pool.
func TestSubmitWithContext_Success(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	var executed int32
	var wg sync.WaitGroup
	wg.Add(1)
	err := p.SubmitWithContext(context.Background(), func() {
		atomic.AddInt32(&executed, 1)
		wg.Done()
	})
	require.NoError(t, err)
	wg.Wait()
	require.Equal(t, int32(1), executed)
	stats := p.GetStats()
	require.Equal(t, int64(1), stats["total_tasks"].(int64))
}

// TestDefaultWorkerPool_ShutdownWithTimeout validates that the worker pool shuts down within the specified timeout duration.
func TestDefaultWorkerPool_ShutdownWithTimeout(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	var taskStarted sync.WaitGroup
	taskStarted.Add(1)
	p.Submit(func() {
		taskStarted.Done()
		time.Sleep(500 * time.Millisecond)
	})
	taskStarted.Wait()
	start := time.Now()
	err := p.ShutdownWithTimeout(100 * time.Millisecond)
	elapsed := time.Since(start)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrShutdownTimeout)
	require.True(t, elapsed >= 100*time.Millisecond)
	require.True(t, elapsed < 200*time.Millisecond)
}

// TestShutdownWithTimeout_Success verifies that ShutdownWithTimeout successfully shuts down the worker pool within the timeout period.
func TestShutdownWithTimeout_Success(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	err := p.ShutdownWithTimeout(100 * time.Millisecond)
	require.NoError(t, err)
	require.True(t, p.IsShutdown())
}

// TestSetWorkerCount_Increase verifies that increasing the worker count adjusts the pool and handles concurrent tasks correctly.
func TestSetWorkerCount_Increase(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	stats := p.GetStats()
	require.Equal(t, int32(1), stats["worker_count"].(int32))
	p.SetWorkerCount(3)
	stats = p.GetStats()
	require.Equal(t, int32(3), stats["worker_count"].(int32))
	var executed int32
	var wg sync.WaitGroup
	taskCount := 3
	wg.Add(taskCount)
	for i := 0; i < taskCount; i++ {
		p.Submit(func() {
			atomic.AddInt32(&executed, 1)
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		})
	}
	wg.Wait()
	require.Equal(t, int32(taskCount), executed)
	stats = p.GetStats()
	require.Equal(t, int64(taskCount), stats["total_tasks"].(int64))
}

// TestSetWorkerCount_Decrease verifies that decreasing the worker count in the pool dynamically adjusts the active workers.
func TestSetWorkerCount_Decrease(t *testing.T) {
	t.Parallel()
	p := NewDefaultWorkerPool(
		context.Background(),
		WithWorkerCount(4),
		WithLogPath(os.DevNull),
	)
	defer p.Shutdown()
	stats := p.GetStats()
	require.Equal(t, int32(4), stats["worker_count"].(int32))
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		p.Submit(func() {
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		})
	}
	p.SetWorkerCount(2)
	stats = p.GetStats()
	require.Equal(t, int32(2), stats["worker_count"].(int32))
	wg.Wait()
	var executed int32
	var wg2 sync.WaitGroup
	wg2.Add(1)
	p.Submit(func() {
		atomic.AddInt32(&executed, 1)
		wg2.Done()
	})
	wg2.Wait()
	require.Equal(t, int32(1), executed)
}
