package models

import (
	"github.com/google/uuid"
	"time"
)

type ReturnEntry struct {
	OrderID    uuid.UUID `json:"order_id"`
	UserID     uuid.UUID `json:"user_id"`
	ReturnedAt time.Time `json:"returned_at"`
}
