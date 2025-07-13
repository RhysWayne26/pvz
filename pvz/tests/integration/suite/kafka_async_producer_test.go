//go:build integration || e2e

package suite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"
	brokersPkg "pvz-cli/infrastructure/brokers"
)

type KafkaAsyncProducerSuite struct {
	suite.Suite
	deps *kafkaSuiteDeps
}

type kafkaSuiteDeps struct {
	brokers     []string
	topic       string
	admin       sarama.ClusterAdmin
	producer    *brokersPkg.KafkaAsyncProducer
	consumer    sarama.Consumer
	partition   sarama.PartitionConsumer
	testTopics  []string
	topicsMutex sync.Mutex
}

func TestKafkaAsyncProducerSuite(t *testing.T) {
	t.Parallel()
	suite.RunSuite(t, new(KafkaAsyncProducerSuite))
}

func (s *KafkaAsyncProducerSuite) BeforeAll(t provider.T) {
	s.deps = newKafkaSuiteDeps()
}

func (s *KafkaAsyncProducerSuite) AfterAll(t provider.T) {
	s.forceCleanupAllTestTopics()
	s.cleanupResources()
}

func (s *KafkaAsyncProducerSuite) forceCleanupAllTestTopics() {
	if s.deps == nil || s.deps.admin == nil {
		return
	}
	topics, err := s.deps.admin.ListTopics()
	if err != nil {
		log.Printf("failed to list topics: %v\n", err)
		return
	}
	for topicName := range topics {
		if strings.HasPrefix(topicName, "test-") || strings.HasPrefix(topicName, "no-topic-") {
			err := s.deps.admin.DeleteTopic(topicName)
			if err != nil {
				log.Printf("failed to delete topic %s: %v\n", topicName, err)
			} else {
				log.Printf("successfully deleted topic %s\n", topicName)
			}
		}
	}
}

func (s *KafkaAsyncProducerSuite) TestIntegration(t provider.T) {
	t.WithNewStep("Happy path: send & receive", func(ctx provider.StepCtx) {
		topic := s.createTestTopic("integration")
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		payload := []byte(`{"event":"test","value":52}`)
		require.NoError(t, s.deps.producer.Send(c, topic, payload), "send should succeed")
		time.Sleep(500 * time.Millisecond)
		partition := s.createPartitionConsumer(topic)
		defer partition.Close()
		select {
		case msg := <-partition.Messages():
			require.Equal(t, payload, msg.Value)
		case err := <-partition.Errors():
			t.Fatalf("consumer error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for message")
		}
	})
}

func (s *KafkaAsyncProducerSuite) TestBrokerDown(t provider.T) {
	t.WithNewStep("Producer on dead broker should error", func(ctx provider.StepCtx) {
		prod, err := brokersPkg.NewKafkaAsyncProducer([]string{"127.0.0.1:9999"})
		if err != nil {
			t.Logf("Producer creation failed as expected: %v", err)
			return
		}
		defer prod.Close()
		c, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err = prod.Send(c, "whatever", []byte("payload"))
		require.Error(t, err, "expect error trying to send to dead broker")
	})
}

func (s *KafkaAsyncProducerSuite) TestNonExistentTopic(t provider.T) {
	t.WithNewStep("Send to missing topic — async enqueue OK", func(ctx provider.StepCtx) {
		missing := fmt.Sprintf("no-topic-%d", time.Now().UnixNano())
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := s.deps.producer.Send(c, missing, []byte("hello"))
		require.NoError(t, err, "enqueue to dead topic should not fall")
		time.Sleep(1 * time.Second)
		s.deleteTestTopic(missing)
	})
}

func (s *KafkaAsyncProducerSuite) TestSendTimeout(t provider.T) {
	t.WithNewStep("Very short deadline — either enqueue, or either DeadlineExceeded", func(ctx provider.StepCtx) {
		c, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		err := s.deps.producer.Send(c, s.deps.topic, []byte("ok"))
		require.True(t,
			err == nil || errors.Is(err, context.DeadlineExceeded),
			"expected successful enqueue or DeadlineExceeded, but got %v", err,
		)
	})
}

func (s *KafkaAsyncProducerSuite) TestContextCancellation(t provider.T) {
	t.WithNewStep("Cancel before Send — either Canceled, or either nil", func(ctx provider.StepCtx) {
		c, cancel := context.WithCancel(context.Background())
		cancel()

		err := s.deps.producer.Send(c, s.deps.topic, []byte("nope"))
		require.True(t,
			err == nil || errors.Is(err, context.Canceled),
			"expected Cancelled or nil, but got %v", err,
		)
	})
}

func (s *KafkaAsyncProducerSuite) TestEmptyMessage(t provider.T) {
	t.WithNewStep("Empty payload", func(ctx provider.StepCtx) {
		topic := s.createTestTopic("empty")
		partition := s.createPartitionConsumer(topic)
		defer partition.Close()

		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		require.NoError(t, s.deps.producer.Send(c, topic, []byte{}))
		select {
		case msg := <-partition.Messages():
			require.Empty(t, msg.Value)
		case <-c.Done():
			t.Fatal("did not get empty message")
		}
	})
}

func (s *KafkaAsyncProducerSuite) TestConcurrentSend(t provider.T) {
	t.WithNewStep("Parallel N messages sending", func(ctx provider.StepCtx) {
		topic := s.createTestTopic("concurrent")
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		const n = 20
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func(i int) {
				defer wg.Done()
				p := []byte(fmt.Sprintf("msg-%d", i))
				require.NoError(t, s.deps.producer.Send(c, topic, p))
			}(i)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
		partition := s.createPartitionConsumer(topic)
		defer partition.Close()
		seen := make(map[string]bool)
		deadline := time.After(5 * time.Second)
	loop:
		for {
			select {
			case m := <-partition.Messages():
				seen[string(m.Value)] = true
				if len(seen) == n {
					break loop
				}
			case <-deadline:
				t.Fatalf("expected %d msgs, got %d", n, len(seen))
			}
		}
	})
}

func (s *KafkaAsyncProducerSuite) TestSendWithKey(t provider.T) {
	t.WithNewStep("SendWithKey – check key and value", func(ctx provider.StepCtx) {
		topic := s.createTestTopic("withkey")

		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		key := []byte("the-key")
		val := []byte("with-key")
		require.NoError(t, s.deps.producer.SendWithKey(c, topic, key, val))
		time.Sleep(500 * time.Millisecond)
		partition := s.createPartitionConsumer(topic)
		defer partition.Close()
		select {
		case msg := <-partition.Messages():
			require.Equal(t, key, msg.Key)
			require.Equal(t, val, msg.Value)
		case <-time.After(3 * time.Second):
			t.Fatal("timeout keyed message")
		}
	})
}

func (s *KafkaAsyncProducerSuite) deleteTestTopic(topic string) {
	err := s.deps.admin.DeleteTopic(topic)
	if err != nil {
		log.Printf("failed to delete topic %s: %v\n", topic, err)
	} else {
		log.Printf("successfully deleted topic %s\n", topic)
	}
}

func (s *KafkaAsyncProducerSuite) createTestTopic(testName string) string {
	topic := fmt.Sprintf("test-%s-%d", testName, time.Now().UnixNano())
	td := &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}
	err := s.deps.admin.CreateTopic(topic, td, false)
	if err != nil && err != sarama.ErrTopicAlreadyExists {
		panic(fmt.Sprintf("failed to create test topic %s: %v", topic, err))
	}
	time.Sleep(200 * time.Millisecond)
	return topic
}

func (s *KafkaAsyncProducerSuite) createPartitionConsumer(topic string) sarama.PartitionConsumer {
	var (
		partition sarama.PartitionConsumer
		err       error
	)
	for i := 0; i < 5; i++ {
		partition, err = s.deps.consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		panic(fmt.Sprintf("failed to create partition consumer for %s: %v", topic, err))
	}
	return partition
}

func newKafkaSuiteDeps() *kafkaSuiteDeps {
	brokers := []string{"localhost:9094"}
	topic := fmt.Sprintf("test-topic-%d", time.Now().UnixNano())
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_0_0_0
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create admin: %v", err))
	}
	_ = admin.DeleteTopic(topic)
	time.Sleep(100 * time.Millisecond)
	td := &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}
	err = admin.CreateTopic(topic, td, false)
	if err != nil && err != sarama.ErrTopicAlreadyExists {
		panic(fmt.Sprintf("failed to create topic: %v", err))
	}
	prod, err := brokersPkg.NewKafkaAsyncProducer(brokers)
	if err != nil {
		panic(fmt.Sprintf("failed to create producer: %v", err))
	}
	consCfg := sarama.NewConfig()
	consCfg.Version = sarama.V3_0_0_0
	consCfg.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer(brokers, consCfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create consumer: %v", err))
	}
	part, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(fmt.Sprintf("failed to create partition consumer: %v", err))
	}

	return &kafkaSuiteDeps{
		brokers:     brokers,
		topic:       topic,
		admin:       admin,
		producer:    prod,
		consumer:    consumer,
		partition:   part,
		testTopics:  make([]string, 0),
		topicsMutex: sync.Mutex{},
	}
}

func (s *KafkaAsyncProducerSuite) cleanupResources() {
	if s.deps == nil {
		return
	}
	if s.deps.partition != nil {
		_ = s.deps.partition.Close()
	}
	if s.deps.consumer != nil {
		_ = s.deps.consumer.Close()
	}
	if s.deps.producer != nil {
		_ = s.deps.producer.Close()
	}
	if s.deps.admin != nil {
		log.Printf("deleting main topic: %s\n", s.deps.topic)
		_ = s.deps.admin.DeleteTopic(s.deps.topic)
		s.deps.topicsMutex.Lock()
		log.Printf("found %d test topics to delete\n", len(s.deps.testTopics))
		for _, topic := range s.deps.testTopics {
			log.Printf("deleting test topic: %s\n", topic)
			err := s.deps.admin.DeleteTopic(topic)
			if err != nil {
				log.Printf("failed to delete topic %s: %v\n", topic, err)
			} else {
				log.Printf("deleted topic %s\n", topic)
			}
		}
		s.deps.topicsMutex.Unlock()
		_ = s.deps.admin.Close()
	}
}
