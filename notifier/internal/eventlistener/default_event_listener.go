package eventlistener

import (
	"context"
	"github.com/IBM/sarama"
	"log/slog"
	"notifier/internal/infrastructure/brokers"
)

var _ EventListener = (*DefaultEventListener)(nil)

// DefaultEventListener uses a KafkaConsumer to process messages from specific topics with a provided handler.
type DefaultEventListener struct {
	consumer brokers.KafkaConsumer
	topics   []string
	handler  sarama.ConsumerGroupHandler
}

// NewDefaultEventListener creates a new DefaultEventListener using the provided KafkaConsumer, topics, and handler.
func NewDefaultEventListener(consumer brokers.KafkaConsumer, topics []string, handler sarama.ConsumerGroupHandler) *DefaultEventListener {
	return &DefaultEventListener{
		consumer: consumer,
		topics:   topics,
		handler:  handler,
	}
}

// Listen starts consuming messages from the specified Kafka topics using the provided context and message handler.
func (l *DefaultEventListener) Listen(ctx context.Context) error {
	return l.consumer.Receive(ctx, l.topics, l.handler)
}

// Stop terminates the Kafka consumer and releases associated resources, logging any errors encountered during closure.
func (l *DefaultEventListener) Stop() {
	err := l.consumer.Close()
	if err != nil {
		slog.Error("error closing kafka consumer", "error", err)
	}
}
