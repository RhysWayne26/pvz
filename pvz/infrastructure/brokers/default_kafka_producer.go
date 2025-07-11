package brokers

import (
	"context"
	"github.com/segmentio/kafka-go"
	"sync"
	"time"
)

var _ KafkaProducer = (*DefaultKafkaProducer)(nil)

type DefaultKafkaProducer struct {
	brokers []string
	writers sync.Map
}

func NewDefaultProducer(brokers []string) *DefaultKafkaProducer {
	return &DefaultKafkaProducer{
		brokers: brokers,
	}
}

func (p *DefaultKafkaProducer) Send(ctx context.Context, topic string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	w := p.getOrCreateWriter(topic)
	msg := kafka.Message{
		Value: payload,
		Time:  nowUTC(),
	}
	return w.WriteMessages(ctx, msg)
}

func (p *DefaultKafkaProducer) SendWithKey(ctx context.Context, topic string, key []byte, payload []byte) error {
	writer := p.getOrCreateWriter(topic)
	msg := kafka.Message{
		Key:   key,
		Value: payload,
		Time:  nowUTC(),
	}
	return writer.WriteMessages(ctx, msg)
}

func (p *DefaultKafkaProducer) Close() error {
	var firstErr error
	p.writers.Range(func(_ interface{}, w interface{}) bool {
		writer := w.(*kafka.Writer)
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
}

func (p *DefaultKafkaProducer) getOrCreateWriter(topic string) *kafka.Writer {
	if wi, ok := p.writers.Load(topic); ok {
		return wi.(*kafka.Writer)
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		Async:        false,
	}
	p.writers.Store(topic, w)
	return w
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
