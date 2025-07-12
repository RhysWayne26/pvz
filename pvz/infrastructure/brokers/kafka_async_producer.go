package brokers

import (
	"context"
	"github.com/IBM/sarama"
	"log"
)

var _ KafkaProducer = (*KafkaAsyncProducer)(nil)

// KafkaAsyncProducer wraps a Sarama AsyncProducer to provide non-blocking Kafka message publishing.
type KafkaAsyncProducer struct {
	prod sarama.AsyncProducer
}

// NewKafkaAsyncProducer creates a new asynchronous Kafka producer configured with the provided broker addresses. .
func NewKafkaAsyncProducer(brokers []string) (*KafkaAsyncProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case m, ok := <-p.Successes():
				if !ok {
					p = nil
					return
				}
				log.Printf("message sent to %s/%d", m.Topic, m.Partition)
			case err, ok := <-p.Errors():
				if !ok {
					return
				}
				log.Printf("failed to send message: %v", err)
			}
		}
	}()

	return &KafkaAsyncProducer{prod: p}, nil
}

// Send publishes a message to the specified Kafka topic asynchronously, handling context cancellation if applicable.
func (p *KafkaAsyncProducer) Send(ctx context.Context, topic string, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.prod.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}:
		return nil
	}
}

// SendWithKey sends a message with a specified key to the Kafka topic asynchronously.
func (p *KafkaAsyncProducer) SendWithKey(ctx context.Context, topic string, key, payload []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.prod.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}:
		return nil
	}
}

// Close terminates the underlying Kafka asynchronous producer, ensuring all outgoing messages are flushed or discarded.
func (p *KafkaAsyncProducer) Close() error {
	return p.prod.Close()
}
