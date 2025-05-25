package requests

type ListOrdersFilter struct {
	UserID string
	InPvz  *bool
	LastID string
	Page   *int
	Limit  *int
	Last   *int
}
