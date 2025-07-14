package brokers

import (
	"context"
	"github.com/IBM/sarama"
	"log/slog"
	"sync"
)

var _ KafkaConsumer = (*DefaultKafkaConsumer)(nil)

// DefaultKafkaConsumer represents a Kafka consumer group wrapper with internal context and wait group management.
// It encapsulates a sarama.ConsumerGroup, a context cancel function, and a sync.WaitGroup for operations coordination.
type DefaultKafkaConsumer struct {
	cg     sarama.ConsumerGroup
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDefaultKafkaConsumer initializes a new Kafka consumer group with the provided brokers and group ID configuration.
func NewDefaultKafkaConsumer(brokers []string, groupID string) (KafkaConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_0_0_0
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Return.Errors = true
	cg, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &DefaultKafkaConsumer{cg: cg}, nil
}

// Receive starts consuming messages from the specified Kafka topics using the provided context and consumer group handler.
func (c *DefaultKafkaConsumer) Receive(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	ctx, c.cancel = context.WithCancel(ctx)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.cg.Consume(ctx, topics, handler); err != nil {
				slog.Error("[KafkaConsumer] consume error", "error", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case err := <-c.cg.Errors():
				if err != nil {
					slog.Error("KafkaConsumer group error", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (c *DefaultKafkaConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.cg.Close()
}
