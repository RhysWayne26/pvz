package repositories

import (
	"context"
	"pvz-cli/internal/models"
	"time"
)

var _ OutboxRepository = (*NoOpOutboxRepository)(nil)

// NoOpOutboxRepository is a no-operation implementation of the OutboxRepository interface with all methods returning nil.
type NoOpOutboxRepository struct{}

func NewNoOpOutboxRepository() *NoOpOutboxRepository {
	return &NoOpOutboxRepository{}
}

// FetchPending retrieves a list of pending outbox events up to the specified limit and older than the given retry delay.
func (r *NoOpOutboxRepository) FetchPending(ctx context.Context, limit int, retryDelay time.Duration) ([]models.OutboxEvent, error) {
	return nil, nil
}

// SetProcessing marks an event as being processed in the no-operation outbox repository implementation.
func (r *NoOpOutboxRepository) SetProcessing(ctx context.Context, eventID uint64) error {
	return nil
}

// SetCompleted marks the specified event as completed using its ID and the timestamp when it was sent.
func (r *NoOpOutboxRepository) SetCompleted(ctx context.Context, eventID uint64, sentAt time.Time) error {
	return nil
}

// SetFailed marks an event as failed in the no-operation outbox repository implementation with the given error message.
func (r *NoOpOutboxRepository) SetFailed(ctx context.Context, eventID uint64, errMsg string) error {
	return nil
}

// Create inserts a new outbox event with the provided payload in the no-operation outbox repository implementation.
func (r *NoOpOutboxRepository) Create(ctx context.Context, payload []byte) error {
	return nil
}
