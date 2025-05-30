package requests

// ClientReturnsRequest contains parameters for processing client returns
type ClientReturnsRequest struct {
	OrderIDs []string
	UserID   string
}
