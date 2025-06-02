package models

import (
	"time"
)

// ReturnEntry represents a record of order return by client operation
type ReturnEntry struct {
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	ReturnedAt time.Time `json:"returned_at"`
}
