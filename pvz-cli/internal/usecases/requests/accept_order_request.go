package requests

import (
	"time"
)

type AcceptOrderRequest struct {
	OrderID   string
	UserID    string
	ExpiresAt time.Time
}
