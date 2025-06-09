package params

// AcceptOrderParams contains parameters for accept-order command
type AcceptOrderParams struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ExpiresAt string `json:"expires_at"`
	Weight    string `json:"weight"`
	Price     string `json:"price"`
	Package   string `json:"package"`
}

// ReturnOrderParams contains parameters for return-order command
type ReturnOrderParams struct {
	OrderID string `json:"order_id"`
}

// ProcessOrdersParams contains parameters for process-orders command
type ProcessOrdersParams struct {
	UserID   string `json:"user_id"`
	Action   string `json:"action"`
	OrderIDs string `json:"order_ids"`
}

// ListOrdersParams contains parameters for list-orders command
type ListOrdersParams struct {
	UserID string `json:"user_id"`
	InPvz  *bool  `json:"in_pvz,omitempty"`
	Last   *int   `json:"last,omitempty"`
	LastID string `json:"last_id,omitempty"`
	Page   *int   `json:"page,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}

// ListReturnsParams contains parameters for list-returns command
type ListReturnsParams struct {
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

// ScrollOrdersParams contains parameters for scroll-orders command
type ScrollOrdersParams struct {
	UserID string `json:"user_id"`
	Limit  *int   `json:"limit,omitempty"`
	LastID string `json:"last_id,omitempty"`
}

// ImportOrdersParams contains parameters for import-orders command
type ImportOrdersParams struct {
	File string `json:"file"`
}
