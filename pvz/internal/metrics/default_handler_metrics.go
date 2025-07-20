package metrics

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
)

// DefaultHandlerMetrics tracks metrics for handler performance, such as the total number of orders served.
type DefaultHandlerMetrics struct {
	ordersServed prometheus.Counter
}

// NewDefaultHandlerMetrics creates and registers default handler metrics for tracking system performance.
func NewDefaultHandlerMetrics(reg prometheus.Registerer) (*DefaultHandlerMetrics, error) {
	ordersServed := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "pvz",
		Subsystem: "orders",
		Name:      "served_total",
		Help:      "Total number of orders served by system",
	})

	if err := reg.Register(ordersServed); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			ordersServed = are.ExistingCollector.(prometheus.Counter)
		}
	}

	return &DefaultHandlerMetrics{
		ordersServed: ordersServed,
	}, nil
}

// IncOrdersServed increments the ordersServed counter by the specified delta value.
func (m *DefaultHandlerMetrics) IncOrdersServed(delta float64) {
	m.ordersServed.Add(delta)
}
