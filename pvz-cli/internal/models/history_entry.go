package models

import (
	"github.com/google/uuid"
	"time"
)

type HistoryEntry struct {
	OrderID   uuid.UUID `json:"order_id"`
	Event     EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}
