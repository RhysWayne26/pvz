package models

import (
	"time"
)

// ReturnEntry represents a record of order return by client operation
type ReturnEntry struct {
	OrderID    uint64    `json:"order_id"`
	UserID     uint64    `json:"user_id"`
	ReturnedAt time.Time `json:"returned_at"`
}
