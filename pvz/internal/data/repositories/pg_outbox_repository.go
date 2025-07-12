package repositories

import (
	"context"
	"fmt"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/data/queries"
	"pvz-cli/internal/models"
	"time"
)

var _ OutboxRepository = (*PGOutboxRepository)(nil)

type PGOutboxRepository struct {
	client db.PGXClient
}

func NewPGOutboxRepository(client db.PGXClient) *PGOutboxRepository {
	return &PGOutboxRepository{
		client: client,
	}
}

func (r *PGOutboxRepository) Create(ctx context.Context, eventID uint64, payload []byte) error {
	_, err := r.client.ExecCtx(
		ctx,
		db.WriteMode,
		queries.CreateOutboxEventSQL,
		eventID,
		payload,
	)
	if err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}
	return nil
}

func (r *PGOutboxRepository) FetchPending(ctx context.Context, limit int, retryDelay time.Duration) ([]models.OutboxEvent, error) {
	rows, err := r.client.QueryCtx(
		ctx,
		db.ReadMode,
		queries.FetchPendingSQL,
		models.OutboxStatusCreated,
		int(retryDelay.Seconds()),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch pending events: %w", err)
	}
	defer rows.Close()
	events := make([]models.OutboxEvent, 0, limit)
	for rows.Next() {
		var e models.OutboxEvent
		if err := rows.Scan(
			&e.EventID,
			&e.Payload,
			&e.Status,
			&e.Error,
			&e.CreatedAt,
			&e.SentAt,
			&e.Attempts,
			&e.LastAttemptAt,
		); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, e)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate outbox rows: %w", rows.Err())
	}
	return events, nil
}

func (r *PGOutboxRepository) SetProcessing(ctx context.Context, eventID uint64) error {
	_, err := r.client.ExecCtx(
		ctx,
		db.WriteMode,
		queries.SetProcessingSQL,
		eventID,
	)
	if err != nil {
		return fmt.Errorf("mark processing outbox event: %w", err)
	}
	return nil
}

func (r *PGOutboxRepository) SetCompleted(ctx context.Context, eventID uint64, sentAt time.Time) error {
	_, err := r.client.ExecCtx(
		ctx,
		db.WriteMode,
		queries.SetCompletedSQL,
		eventID,
		sentAt,
	)
	if err != nil {
		return fmt.Errorf("mark completed outbox event: %w", err)
	}
	return nil
}

func (r *PGOutboxRepository) SetFailed(ctx context.Context, eventID uint64, errMsg string) error {
	_, err := r.client.ExecCtx(
		ctx,
		db.WriteMode,
		queries.SetFailedSQL,
		eventID,
		errMsg,
	)
	if err != nil {
		return fmt.Errorf("mark failed outbox event: %w", err)
	}
	return nil
}
