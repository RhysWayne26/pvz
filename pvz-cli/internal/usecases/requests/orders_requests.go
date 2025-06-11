package requests

import (
	"pvz-cli/internal/models"
	"time"
)

// ProcessAction defines the type of operation to perform on orders.
type ProcessAction string

const (
	// ActionIssue represents issue request
	ActionIssue ProcessAction = "issue"
	// ActionReturn represents return from client request
	ActionReturn ProcessAction = "return"
)

// AcceptOrderRequest contains parameters for accepting an order with package pricing
type AcceptOrderRequest struct {
	OrderID   uint64
	UserID    uint64
	ExpiresAt time.Time
	Weight    float32
	Price     float32
	Package   models.PackageType
}

// ReturnOrderRequest contains parameters for returning an order to courier
type ReturnOrderRequest struct {
	OrderID uint64
}

// ProcessOrdersRequest aggregates user ID, list of order IDs, and the action to be performed.
type ProcessOrdersRequest struct {
	UserID   uint64
	OrderIDs []uint64
	Action   ProcessAction
}

// IssueOrdersRequest contains parameters for issuing orders to clients
type IssueOrdersRequest struct {
	OrderIDs []uint64
	UserID   uint64
}

// ClientReturnsRequest contains parameters for processing client returns
type ClientReturnsRequest struct {
	OrderIDs []uint64
	UserID   uint64
}

// ScrollOrdersRequest contains parameters for infinite scroll orders listing
type ScrollOrdersRequest struct {
	UserID uint64
	Limit  *int
	LastID *uint64
}

// ImportOrdersRequest contains a list of accept order request to be performed.
type ImportOrdersRequest struct {
	Statuses []ImportOrderStatus
}

// ImportOrderStatus represents one item in import batch for request
type ImportOrderStatus struct {
	ItemNumber int
	OrderID    uint64
	Request    *AcceptOrderRequest
	Error      error
}
