package common

// ProcessResult represents the outcome of processing a single order operation
type ProcessResult struct {
	OrderID string
	Error   error
}
