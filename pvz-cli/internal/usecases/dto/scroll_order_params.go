package dto

// ScrollOrdersParams contains parameters for scroll-orders command
type ScrollOrdersParams struct {
	UserID string `json:"user_id"`
	Limit  *int   `json:"limit,omitempty"`
}
