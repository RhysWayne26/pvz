package models

import (
	"time"
)

// EventType represents different types of order lifecycle events
type EventType string

// Order lifecycle events
const (
	EventAccepted            EventType = "ACCEPTED"
	EventIssued              EventType = "ISSUED"
	EventReturnedFromClient  EventType = "RETURNED_BY_CLIENT"
	EventReturnedToWarehouse EventType = "RETURNED_TO_WAREHOUSE"
)

// HistoryEntry represents a single event in order lifecycle history
type HistoryEntry struct {
	OrderID   uint64    `json:"order_id"`
	Event     EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}
