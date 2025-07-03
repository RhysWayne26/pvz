package shared

// ProcessResult represents the outcome of processing a single order operation
type ProcessResult struct {
	OrderID uint64
	Error   error
}
