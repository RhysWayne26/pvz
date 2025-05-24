package models

import (
	"time"
)

type Order struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
}
