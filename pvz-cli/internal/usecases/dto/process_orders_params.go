package dto

// ProcessOrdersParams contains parameters for process-orders command
type ProcessOrdersParams struct {
	UserID   string `json:"user_id"`
	Action   string `json:"action"`
	OrderIDs string `json:"order_ids"`
}
