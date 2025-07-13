package workers

import (
	"context"
	"log/slog"
	"pvz-cli/infrastructure/brokers"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/models"
	"time"
)

var _ OutboxDispatcher = (*DefaultOutboxDispatcher)(nil)

type DefaultOutboxDispatcher struct {
	repo         repositories.OutboxRepository
	producer     brokers.KafkaProducer
	topic        string
	batchSize    int
	retryDelay   time.Duration
	pollInterval time.Duration
	cancel       context.CancelFunc
	done         chan struct{}
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
		done:         make(chan struct{}),
	}
}

// Dispatch initiates the outbox dispatcher, starting acquisition and processing loops until the context is canceled.
func (w *DefaultOutboxDispatcher) Dispatch(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	go w.acquireLoop(ctx)
	go w.processLoop(ctx)
	<-ctx.Done()
	close(w.done)
	return ctx.Err()
}

// Stop gracefully stops the dispatcher by canceling the context and waiting for ongoing processes to finish.
func (w *DefaultOutboxDispatcher) Stop() {
	if w.cancel != nil {
		w.cancel()
		<-w.done
	}
}

func (w *DefaultOutboxDispatcher) acquireLoop(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.repo.SetProcessing(ctx, w.batchSize, w.retryDelay); err != nil {
				slog.Error("failed to mark events as processing", "error", err)
			}
		}
	}
}

func (w *DefaultOutboxDispatcher) processLoop(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := w.repo.GetProcessingEvents(ctx, w.batchSize, w.retryDelay)
			if err != nil {
				slog.Error("failed to get processing events", "error", err)
				continue
			}
			for _, event := range events {
				w.dispatchEvent(ctx, event)
			}
		}
	}
}

func (w *DefaultOutboxDispatcher) dispatchEvent(ctx context.Context, ev models.OutboxEvent) (retry bool) {
	if err := w.producer.Send(ctx, w.topic, []byte(ev.Payload)); err != nil {
		if ev.Attempts >= constants.EventSendingMaxAttempts {
			_ = w.repo.SetFailed(ctx, ev.EventID, ErrNoAttemptsLeft)
		} else {
			_ = w.repo.UpdateError(ctx, ev.EventID, err.Error())
		}
		slog.Error("send to Kafka failed", "id", ev.EventID, "attempt", ev.Attempts, "err", err)
		return true
	}

	if err := w.repo.SetCompleted(ctx, ev.EventID, time.Now()); err != nil {
		slog.Error("mark completed failed", "id", ev.EventID, "err", err)
	} else {
		slog.Info("event dispatched", "id", ev.EventID)
	}
	return false
}
