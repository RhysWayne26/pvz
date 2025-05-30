package requests

// ScrollOrdersRequest contains parameters for infinite scroll orders listing
type ScrollOrdersRequest struct {
	UserID  string
	Limit   *int
	AfterID string
}
