package models

type OrderStatus string

const (
	Accepted OrderStatus = "ACCEPTED"
	Returned OrderStatus = "RETURNED"
	Issued   OrderStatus = "ISSUED"
)
