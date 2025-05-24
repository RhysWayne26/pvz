package shutdown

import "sync/atomic"

var flag int32

func Signal() {
	atomic.StoreInt32(&flag, 1)
}

func IsShuttingDown() bool {
	return atomic.LoadInt32(&flag) == 1
}
