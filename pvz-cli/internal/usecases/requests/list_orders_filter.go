package requests

// ListOrdersFilter contains filtering and pagination parameters for listing orders
type ListOrdersFilter struct {
	UserID string
	InPvz  *bool
	LastID string
	Page   *int
	Limit  *int
	Last   *int
}
