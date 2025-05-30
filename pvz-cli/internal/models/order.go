package models

import (
	"time"
)

type Order struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   *time.Time  `json:"issued_at,omitempty"`
	ReturnedAt *time.Time  `json:"returned_at,omitempty"`
	Package    PackageType `json:"package"`
	Weight     float64     `json:"weight"`
	Price      float64     `json:"price"`
}
