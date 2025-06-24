package models

import (
	"time"
)

// EventType represents different types of order lifecycle events
type EventType int32

// Order lifecycle events
const (
	EventAccepted            EventType = 1
	EventIssued              EventType = 2
	EventReturnedByClient    EventType = 3
	EventReturnedToWarehouse EventType = 4
)

// HistoryEntry represents a single event in order lifecycle history
type HistoryEntry struct {
	OrderID   uint64    `json:"order_id" db:"order_id"`
	Event     EventType `json:"event_type" db:"event"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

func (e EventType) String() string {
	switch e {
	case EventAccepted:
		return "ACCEPTED"
	case EventIssued:
		return "ISSUED"
	case EventReturnedByClient:
		return "RETURNED_BY_CLIENT"
	case EventReturnedToWarehouse:
		return "RETURNED_TO_WAREHOUSE"
	default:
		return "UNKNOWN"
	}
}
