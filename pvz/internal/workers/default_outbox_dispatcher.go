package workers

import (
	"context"
	"log/slog"
	"pvz-cli/infrastructure/brokers"
	"pvz-cli/internal/data/repositories"
	"time"
)

var _ OutboxDispatcher = (*DefaultOutboxDispatcher)(nil)

// DefaultOutboxDispatcher is a default implementation of the OutboxDispatcher interface for handling message dispatching.
type DefaultOutboxDispatcher struct {
	repo         repositories.OutboxRepository
	producer     brokers.KafkaProducer
	topic        string
	batchSize    int
	retryDelay   time.Duration
	pollInterval time.Duration
	cancel       context.CancelFunc
}

// NewDefaultOutboxDispatcher creates and returns a new DefaultOutboxDispatcher with the specified configuration and dependencies.
func NewDefaultOutboxDispatcher(
	repo repositories.OutboxRepository,
	producer brokers.KafkaProducer,
	topic string,
	batchSize int,
	retryDelay time.Duration,
	pollInterval time.Duration,
) *DefaultOutboxDispatcher {
	return &DefaultOutboxDispatcher{
		repo:         repo,
		producer:     producer,
		topic:        topic,
		batchSize:    batchSize,
		retryDelay:   retryDelay,
		pollInterval: pollInterval,
	}
}

func (w *DefaultOutboxDispatcher) Dispatch(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			events, err := w.repo.FetchPending(ctx, w.batchSize, w.retryDelay)
			if err != nil {
				slog.Error("failed to fetch pending events", "error", err)
				time.Sleep(w.retryDelay)
				continue
			}

			for _, event := range events {
				if err := w.repo.SetProcessing(ctx, event.EventID); err != nil {
					slog.Error("failed to mark event as processing", "event_id", event.EventID, "error", err)
					continue
				}

				if err := w.producer.Send(ctx, w.topic, []byte(event.Payload)); err != nil {
					_ = w.repo.SetFailed(ctx, event.EventID, err.Error())
					slog.Error("failed to send event to Kafka", "event_id", event.EventID, "error", err)
					time.Sleep(w.retryDelay)
					continue
				}

				if err := w.repo.SetCompleted(ctx, event.EventID, time.Now()); err != nil {
					slog.Error("failed to set event as completed", "event_id", event.EventID, "error", err)
				}
				slog.Info("event successfully dispatched", "event_id", event.EventID)
			}
		}
	}
}

func (w *DefaultOutboxDispatcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
