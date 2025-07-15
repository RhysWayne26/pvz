package brokers

import "context"

var _ KafkaProducer = (*KafkaNoOpProducer)(nil)

type KafkaNoOpProducer struct{}

func NewKafkaNoOpProducer() *KafkaNoOpProducer {
	return &KafkaNoOpProducer{}
}

func (k KafkaNoOpProducer) Send(ctx context.Context, topic string, payload []byte) error {
	return nil
}

func (k KafkaNoOpProducer) SendWithKey(ctx context.Context, topic string, key []byte, payload []byte) error {
	return nil
}

func (k KafkaNoOpProducer) Close() error {
	return nil
}
