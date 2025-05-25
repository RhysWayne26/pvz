package models

import (
	"time"
)

type HistoryEntry struct {
	OrderID   string    `json:"order_id"`
	Event     EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}
