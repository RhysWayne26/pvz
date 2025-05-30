package requests

import (
	"pvz-cli/internal/models"
	"time"
)

// AcceptOrderRequest contains parameters for accepting an order with package pricing
type AcceptOrderRequest struct {
	OrderID   string
	UserID    string
	ExpiresAt time.Time
	Weight    float64
	Price     float64
	Package   models.PackageType
}
