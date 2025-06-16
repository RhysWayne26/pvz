package responses

import (
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// AcceptOrderResponse represents the result of successfully accepting an order.
type AcceptOrderResponse struct {
	OrderID uint64
	Package models.PackageType
	Price   float32
}

// ReturnOrderResponse represents a successful order return operation.
type ReturnOrderResponse struct {
	OrderID uint64
}

// ProcessOrdersResponse aggregates the results of a batch operation on orders.
type ProcessOrdersResponse struct {
	Processed []uint64
	Failed    []ProcessFailReport
}

// ProcessFailReport represents a failed order processing entry with error.
type ProcessFailReport struct {
	OrderID uint64
	Error   error
}

// ListOrdersResponse represents a list of orders, total count and pagination metadata.
type ListOrdersResponse struct {
	Orders []models.Order
	NextID *uint64
	Total  *int
}

// OrderHistoryResponse contains a list of order history entries.
type OrderHistoryResponse struct {
	History []models.HistoryEntry
}

// ImportOrdersResponse represents the result of an import-orders operation.
type ImportOrdersResponse struct {
	Imported int
	Statuses []requests.ImportOrderStatus
}
