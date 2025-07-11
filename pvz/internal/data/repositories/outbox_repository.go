package repositories

import (
	"context"
	"pvz-cli/internal/models"
	"time"
)

// OutboxRepository defines methods for managing and processing events in the outbox for reliable message delivery.
type OutboxRepository interface {
	Create(ctx context.Context, payload []byte) error
	FetchPending(ctx context.Context, limit int, retryDelay time.Duration) ([]models.OutboxEvent, error)
	SetProcessing(ctx context.Context, eventID uint64) error
	SetCompleted(ctx context.Context, eventID uint64, sentAt time.Time) error
	SetFailed(ctx context.Context, eventID uint64, errMsg string) error
}
