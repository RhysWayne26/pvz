package models

import (
	"github.com/google/uuid"
	"time"
)

type Order struct {
	OrderID    uuid.UUID   `json:"order_id"`
	UserID     uuid.UUID   `json:"user_id"`
	Status     OrderStatus `json:"status"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
}
