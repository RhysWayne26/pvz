package requests

import (
	"github.com/google/uuid"
	"time"
)

type AcceptOrderRequest struct {
	OrderID   uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}
