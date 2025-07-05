package workerpool

// PoolError represents a specific type of error occurring in a pool implementation; used to denote specific pool-related issues like shutdown or timeout errors.
type PoolError string

func (e PoolError) Error() string { return string(e) }

const (

	// ErrPoolShuttingDown is returned when an operation is attempted on a worker pool in the process of shutting down.
	ErrPoolShuttingDown PoolError = "worker pool is shutting down"

	// ErrShutdownTimeout indicates that the worker pool shutdown exceeded the specified timeout duration.
	ErrShutdownTimeout PoolError = "shutdown timed out"
)
