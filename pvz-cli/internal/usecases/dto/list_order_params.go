package dto

// ListOrdersParams contains parameters for list-orders command
type ListOrdersParams struct {
	UserID string `json:"user_id"`
	InPvz  *bool  `json:"in_pvz,omitempty"`
	Last   *int   `json:"last,omitempty"`
	LastID string `json:"last_id,omitempty"`
	Page   *int   `json:"page,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}
