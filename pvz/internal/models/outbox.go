package models

import "time"

// OutboxStatus represents the status of an outbox event in the processing lifecycle.
type OutboxStatus int32

const (
	OutboxStatusCreated    OutboxStatus = 1
	OutboxStatusProcessing OutboxStatus = 2
	OutboxStatusCompleted  OutboxStatus = 3
	OutboxStatusFailed     OutboxStatus = 4
)

// ActorType represents a string-based designation for different types of actors in the system.
type ActorType string

const (
	ActorCourier ActorType = "courier"
	ActorClient  ActorType = "client"
)

func (s OutboxStatus) String() string {
	switch s {
	case OutboxStatusCreated:
		return "CREATED"
	case OutboxStatusProcessing:
		return "PROCESSING"
	case OutboxStatusCompleted:
		return "COMPLETED"
	case OutboxStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// OutboxEvent represents an event stored in the outbox table for eventual processing and delivery.
type OutboxEvent struct {
	EventID       uint64       `db:"id"`
	OrderID       uint64       `db:"order_id"`
	Payload       string       `db:"payload"`
	Status        OutboxStatus `db:"status"`
	Error         string       `db:"error"`
	CreatedAt     time.Time    `db:"created_at"`
	SentAt        *time.Time   `db:"sent_at"`
	Attempts      int          `db:"attempts"`
	LastAttemptAt *time.Time   `db:"last_attempt_at"`
}

type KafkaEvent struct {
	EventID   uint64    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Actor     Actor     `json:"actor"`
	Order     Order     `json:"order"`
	Source    string    `json:"source"`
}

// Actor represents an entity involved in an event, characterized by its type and ID.
type Actor struct {
	Type ActorType `json:"type"`
	ID   uint64    `json:"id"`
}

// MapEventTypeToKafkaEvent maps an EventType to its corresponding Kafka event as a string representation.
func MapEventTypeToKafkaEvent(eventType EventType) string {
	switch eventType {
	case EventAccepted:
		return "order_accepted"
	case EventIssued:
		return "order_issued"
	case EventReturnedByClient:
		return "order_returned_by_client"
	case EventReturnedToWarehouse:
		return "order_returned_to_courier"
	default:
		return "unknown"
	}
}

// MapEventTypeToOrderStatus maps an EventType to its corresponding order status string value.
func MapEventTypeToOrderStatus(eventType EventType) string {
	switch eventType {
	case EventAccepted:
		return "accepted"
	case EventIssued:
		return "issued"
	case EventReturnedByClient:
		return "returned_by_client"
	case EventReturnedToWarehouse:
		return "returned_to_courier"
	default:
		return "unknown"
	}
}
