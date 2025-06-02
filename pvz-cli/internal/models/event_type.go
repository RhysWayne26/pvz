package models

// EventType represents different types of order lifecycle events
type EventType string

// Order lifecycle events
const (
	EventAccepted            EventType = "ACCEPTED"
	EventIssued              EventType = "ISSUED"
	EventReturnedFromClient  EventType = "RETURNED"
	EventReturnedToWarehouse EventType = "ORDER_RETURNED_TO_WAREHOUSE"
)
