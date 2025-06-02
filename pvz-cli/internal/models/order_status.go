package models

// OrderStatus represents the current state of an order in the system
type OrderStatus string

// Available order statuses throughout the order lifecycle
const (
	Accepted OrderStatus = "ACCEPTED"
	Returned OrderStatus = "RETURNED"
	Issued   OrderStatus = "ISSUED"
)
