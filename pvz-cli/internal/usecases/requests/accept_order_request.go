package requests

import (
	"pvz-cli/internal/models"
	"time"
)

type AcceptOrderRequest struct {
	OrderID   string
	UserID    string
	ExpiresAt time.Time
	Weight    float64
	Price     float64
	Package   models.PackageType
}
