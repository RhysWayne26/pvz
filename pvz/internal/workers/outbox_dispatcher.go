package workers

import "context"

// OutboxDispatcher defines the behavior for dispatching outbox messages within a context.
type OutboxDispatcher interface {
	Dispatch(ctx context.Context) error
	Stop()
}
