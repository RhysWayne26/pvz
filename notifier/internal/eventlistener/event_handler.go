package eventlistener

import (
	"fmt"
	"github.com/IBM/sarama"
	"log/slog"
)

type EventHandler struct{}

func NewEventHandler() sarama.ConsumerGroupHandler {
	return &EventHandler{}
}

func (h *EventHandler) Setup(sess sarama.ConsumerGroupSession) error {
	slog.Info("consumer group session setup",
		"generationID", sess.GenerationID(),
		"memberID", sess.MemberID(),
	)
	return nil
}

func (h *EventHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	slog.Info("consumer group session cleanup",
		"generationID", sess.GenerationID(),
		"memberID", sess.MemberID(),
	)
	return nil
}

func (h *EventHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("event received: topic=%s key=%s value=%s\n",
			msg.Topic, string(msg.Key), string(msg.Value),
		)
		sess.MarkMessage(msg, "")
	}
	return nil
}
