package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	brockermocks "pvz-cli/infrastructure/brokers/mocks"
	repmocks "pvz-cli/internal/data/repositories/mocks"
	"pvz-cli/internal/models"
)

func TestNewDefaultOutboxDispatcher(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	assert.NotNil(t, dispatcher)
	assert.Equal(t, "test-topic", dispatcher.topic)
	assert.Equal(t, 10, dispatcher.batchSize)
	assert.Equal(t, time.Second, dispatcher.retryDelay)
	assert.Equal(t, time.Minute, dispatcher.pollInterval)
	assert.NotNil(t, dispatcher.done)
}

func TestDispatchEvent_Success(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	event := models.OutboxEvent{
		EventID:  1,
		Payload:  "test-payload",
		Attempts: 1,
	}
	mockProducer.SendMock.Return(nil)
	mockRepo.SetCompletedMock.Return(nil)
	retry := dispatcher.dispatchEvent(context.Background(), event)
	assert.False(t, retry)
}

func TestDispatchEvent_KafkaFailure_RetryableError(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	event := models.OutboxEvent{
		EventID:  1,
		Payload:  "test-payload",
		Attempts: 1,
	}
	kafkaErr := errors.New("kafka connection failed")
	mockProducer.SendMock.Return(kafkaErr)
	mockRepo.UpdateErrorMock.Return(nil)
	retry := dispatcher.dispatchEvent(context.Background(), event)
	assert.True(t, retry)
}

func TestDispatchEvent_KafkaFailure_MaxAttemptsReached(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	event := models.OutboxEvent{
		EventID:  1,
		Payload:  "test-payload",
		Attempts: 3,
	}
	kafkaErr := errors.New("kafka connection failed")
	mockProducer.SendMock.Return(kafkaErr)
	mockRepo.SetFailedMock.Return(nil)
	retry := dispatcher.dispatchEvent(context.Background(), event)
	assert.True(t, retry)
}

func TestDispatchEvent_SetCompletedFails(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)

	event := models.OutboxEvent{
		EventID:  1,
		Payload:  "test-payload",
		Attempts: 1,
	}
	mockProducer.SendMock.Return(nil)
	mockRepo.SetCompletedMock.Return(errors.New("db error"))
	retry := dispatcher.dispatchEvent(context.Background(), event)
	assert.False(t, retry)
}

func TestStop(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	ctx, cancel := context.WithCancel(context.Background())
	dispatcher.cancel = cancel
	go func() {
		time.Sleep(time.Millisecond * 10)
		dispatcher.Stop()
	}()
	done := make(chan error, 1)
	go func() {
		done <- dispatcher.Dispatch(ctx)
	}()
	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		t.Fatal("Dispatch didn't stop within timeout")
	}
}

func TestDispatch_ContextCanceled(t *testing.T) {
	t.Parallel()
	mockRepo := repmocks.NewOutboxRepositoryMock(t)
	mockProducer := brockermocks.NewKafkaProducerMock(t)
	dispatcher := NewDefaultOutboxDispatcher(
		mockRepo,
		mockProducer,
		"test-topic",
		10,
		time.Second,
		time.Minute,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := dispatcher.Dispatch(ctx)
	assert.Equal(t, context.Canceled, err)
}
