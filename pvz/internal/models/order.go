package models

import (
	"time"
)

// Order represents a package order in the PVZ system
type Order struct {
	OrderID         uint64      `json:"order_id"`
	UserID          uint64      `json:"user_id"`
	Status          OrderStatus `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	ExpiresAt       time.Time   `json:"expires_at"`
	UpdatedStatusAt time.Time   `json:"updated_status_at"`
	Package         PackageType `json:"package"`
	Weight          float32     `json:"weight"`
	Price           float32     `json:"price"`
}

// OrderStatus represents the current state of an order in the system
type OrderStatus string

// Available order statuses throughout the order lifecycle
const (
	Accepted   OrderStatus = "ACCEPTED"
	Returned   OrderStatus = "RETURNED"
	Issued     OrderStatus = "ISSUED"
	Warehoused OrderStatus = "WAREHOUSED"
)

// PackageType represents different types of packaging available for orders
type PackageType string

// Available package types with their weight limits and pricing
const (
	PackageNone    PackageType = "none"     // No packaging (client brings own)
	PackageBag     PackageType = "bag"      // Bag packaging (max 10kg, +5₽)
	PackageBox     PackageType = "box"      // Box packaging (max 30kg, +20₽)
	PackageFilm    PackageType = "film"     // Film packaging (no limit, +1₽)
	PackageBagFilm PackageType = "bag+film" // Bag + film combination
	PackageBoxFilm PackageType = "box+film" // Box + film combination
)
