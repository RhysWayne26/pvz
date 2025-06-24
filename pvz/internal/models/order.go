package models

import (
	"time"
)

// Order represents a package order in the PVZ system
type Order struct {
	OrderID         uint64      `json:"order_id" db:"id"`
	UserID          uint64      `json:"user_id" db:"user_id"`
	Status          OrderStatus `json:"status" db:"status"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	ExpiresAt       time.Time   `json:"expires_at" db:"expires_at"`
	UpdatedStatusAt time.Time   `json:"updated_status_at" db:"updated_status_at"`
	Package         PackageType `json:"package" db:"package"`
	Weight          float32     `json:"weight" db:"weight"`
	Price           float32     `json:"price" db:"price"`
}

// OrderStatus represents the current state of an order in the system
type OrderStatus int32

// Available order statuses throughout the order lifecycle
const (
	Accepted OrderStatus = 1
	Returned OrderStatus = 2
	Issued   OrderStatus = 3
)

// Available order statuses in strings (not for manual use, only for String())
const (
	acceptedStr = "ACCEPTED"
	returnedStr = "RETURNED"
	issuedStr   = "ISSUED"
	unknownStr  = "UNKNOWN"
)

func (s OrderStatus) String() string {
	switch s {
	case Accepted:
		return acceptedStr
	case Returned:
		return returnedStr
	case Issued:
		return issuedStr
	default:
		return unknownStr
	}
}

// PackageType represents different types of packaging available for orders
type PackageType int32

// Available package types with their weight limits and pricing
const (
	PackageNone    PackageType = 0 // No packaging (client brings own)
	PackageBag     PackageType = 1 // Bag packaging (max 10kg, +5₽)
	PackageBox     PackageType = 2 // Box packaging (max 30kg, +20₽)
	PackageFilm    PackageType = 3 // Film packaging (no limit, +1₽)
	PackageBagFilm PackageType = 4 // Bag + film combination
	PackageBoxFilm PackageType = 5 // Box + film combination
)

// Available package types in strings (not for manual use, only for String())
const (
	packageNoneStr    = "none"
	packageBagStr     = "bag"
	packageBoxStr     = "box"
	packageFilmStr    = "film"
	packageBagFilmStr = "bag+film"
	packageBoxFilmStr = "box+film"
	packageUnknownStr = "unknown"
)

func (p PackageType) String() string {
	switch p {
	case PackageNone:
		return packageNoneStr
	case PackageBag:
		return packageBagStr
	case PackageBox:
		return packageBoxStr
	case PackageFilm:
		return packageFilmStr
	case PackageBagFilm:
		return packageBagFilmStr
	case PackageBoxFilm:
		return packageBoxFilmStr
	default:
		return packageUnknownStr
	}
}
