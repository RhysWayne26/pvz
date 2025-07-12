//go:generate minimock -g -i * -o mocks -s "_mock.go"
package brokers

import "context"

type KafkaProducer interface {
	Send(ctx context.Context, topic string, payload []byte) error
	SendWithKey(ctx context.Context, topic string, key []byte, payload []byte) error
	Close() error
}
