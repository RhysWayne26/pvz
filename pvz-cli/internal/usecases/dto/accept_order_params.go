package dto

// AcceptOrderParams contains parameters for accept-order command
type AcceptOrderParams struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ExpiresAt string `json:"expires_at"`
	Weight    string `json:"weight"`
	Price     string `json:"price"`
	Package   string `json:"package"`
}
