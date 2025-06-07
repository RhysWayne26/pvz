package shutdown

import "sync/atomic"

var flag int32

// Signal marks the application as shutting down
func Signal() {
	atomic.StoreInt32(&flag, 1)
}

// IsShuttingDown returns true if the application is in shutdown state
func IsShuttingDown() bool {
	return atomic.LoadInt32(&flag) == 1
}
