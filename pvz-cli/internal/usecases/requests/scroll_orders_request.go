package requests

type ScrollOrdersRequest struct {
	UserID  string
	Limit   *int
	AfterID string
}
