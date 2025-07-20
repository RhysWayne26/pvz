package metrics

// HandlerMetrics defines methods for tracking performance metrics related to handler operations.
type HandlerMetrics interface {
	IncOrdersServed(delta float64)
}
