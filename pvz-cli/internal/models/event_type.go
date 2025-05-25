package models

type EventType string

const (
	EventAccepted            EventType = "ACCEPTED"
	EventIssued              EventType = "ISSUED"
	EventReturnedFromClient  EventType = "RETURNED"
	EventReturnedToWarehouse EventType = "ORDER_RETURNED_TO_WAREHOUSE"
)
