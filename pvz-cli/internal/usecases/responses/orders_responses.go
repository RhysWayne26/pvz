package responses

import (
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
)

// AcceptOrderResponse represents the result of successfully accepting an order.
type AcceptOrderResponse struct {
	OrderID uint64
	Package models.PackageType
	Price   float64
}

// ReturnOrderResponse represents a successful order return operation.
type ReturnOrderResponse struct {
	OrderID uint64
}

// FailedOrder describes the error details for a failed order operation.
type FailedOrder struct {
	Code    apperrors.ErrorCode
	Message string
}

// ProcessOrdersResponse aggregates the results of a batch operation on orders.
type ProcessOrdersResponse struct {
	Processed []uint64
	Failed    map[uint64]FailedOrder
}

// ListOrdersResponse represents a list of orders and the total count.
type ListOrdersResponse struct {
	Orders []models.Order
	Total  int32
}

// ListReturnsResponse represents a list of return entries with pagination metadata.
type ListReturnsResponse struct {
	Returns []models.ReturnEntry
	Page    int
	Limit   int
}

// OrderHistoryResponse contains a list of order history entries.
type OrderHistoryResponse struct {
	History []models.HistoryEntry
}

// FailedImport represents a failed order import with an error code and reason.
type FailedImport struct {
	Code    apperrors.ErrorCode
	Message string
}

// ImportOrdersResponse represents the result of an import-orders operation.
type ImportOrdersResponse struct {
	Imported int32
	Errors   map[uint64]FailedImport
}
