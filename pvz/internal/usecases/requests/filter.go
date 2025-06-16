package requests

import (
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/models"
)

// OrdersFilterRequest defines a flexible filter used by both ListOrders and ListReturns handlers.
type OrdersFilterRequest struct {
	UserID *uint64
	InPvz  *bool
	LastID *uint64
	Page   *int
	Limit  *int
	Last   *int
	Status *models.OrderStatus
}

// NewOrdersFilter creates a new OrdersFilterRequest with default pagination values. Optional modifiers can be applied via functional options.
func NewOrdersFilter(opts ...FilterOption) OrdersFilterRequest {
	var f = OrdersFilterRequest{
		Page:  utils.Ptr(constants.DefaultPage),
		Limit: utils.Ptr(constants.DefaultLimit),
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FilterOption is a functional option that modifies OrdersFilterRequest.
type FilterOption func(*OrdersFilterRequest)

// WithUserID sets the user ID filter.
func WithUserID(id uint64) FilterOption {
	return func(f *OrdersFilterRequest) { f.UserID = utils.Ptr(id) }
}

// WithInPvz sets the in-PVZ filter.
func WithInPvz(inPvz bool) FilterOption {
	return func(f *OrdersFilterRequest) { f.InPvz = utils.Ptr(inPvz) }
}

// WithLastID sets the pagination cursor.
func WithLastID(id uint64) FilterOption {
	return func(f *OrdersFilterRequest) { f.LastID = utils.Ptr(id) }
}

// WithPage sets the page number.
func WithPage(page int) FilterOption {
	return func(f *OrdersFilterRequest) { f.Page = utils.Ptr(page) }
}

// WithLimit sets the page size.
func WithLimit(limit int) FilterOption {
	return func(f *OrdersFilterRequest) { f.Limit = utils.Ptr(limit) }
}

// WithStatus sets the order status filter.
func WithStatus(status models.OrderStatus) FilterOption {
	return func(f *OrdersFilterRequest) { f.Status = utils.Ptr(status) }
}

// WithLast sets the "last N" filter.
func WithLast(last int) FilterOption {
	return func(f *OrdersFilterRequest) { f.Last = utils.Ptr(last) }
}
