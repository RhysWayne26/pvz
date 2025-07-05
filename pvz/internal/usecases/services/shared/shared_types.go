package shared

// BatchEntryProcessedResult represents the result of processing a batch entry, including the OrderID and any associated error.
type BatchEntryProcessedResult struct {
	OrderID uint64
	Error   error
}
