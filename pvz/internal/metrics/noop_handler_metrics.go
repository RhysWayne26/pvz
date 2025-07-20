package metrics

type NoopHandlerMetrics struct{}

func NewNoopHandlerMetrics() (*NoopHandlerMetrics, error) {
	return &NoopHandlerMetrics{}, nil
}

func (m *NoopHandlerMetrics) IncOrdersServed(delta float64) {}
