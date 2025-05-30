package models

import (
	"time"
)

// HistoryEntry represents a single event in order lifecycle history
type HistoryEntry struct {
	OrderID   string    `json:"order_id"`
	Event     EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}
