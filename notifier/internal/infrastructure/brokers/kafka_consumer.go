package brokers

import (
	"context"
	"github.com/IBM/sarama"
)

// KafkaConsumer defines the interface for consuming messages from Kafka topics within a consumer group.
type KafkaConsumer interface {
	Receive(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Close() error
}
