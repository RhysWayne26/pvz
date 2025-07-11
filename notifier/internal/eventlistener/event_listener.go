package eventlistener

import "context"

type EventListener interface {
	Listen(ctx context.Context) error
	Stop()
}
