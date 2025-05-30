package requests

// IssueOrdersRequest contains parameters for issuing orders to clients
type IssueOrdersRequest struct {
	OrderIDs []string
	UserID   string
}
