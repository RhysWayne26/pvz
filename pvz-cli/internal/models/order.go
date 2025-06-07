package models

import (
	"time"
)

// Order represents a package order in the PVZ system
type Order struct {
	OrderID    uint64      `json:"order_id"`
	UserID     uint64      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
	Package    PackageType `json:"package"`
	Weight     float64     `json:"weight"`
	Price      float64     `json:"price"`
}
