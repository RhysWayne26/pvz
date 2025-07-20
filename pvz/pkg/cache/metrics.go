package cache

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Hits      prometheus.Counter
	Misses    prometheus.Counter
	Evictions prometheus.Counter
	Keys      prometheus.Gauge
}

func NewCacheMetrics(namespace, subsystem string, reg prometheus.Registerer) *Metrics {
	hits := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hits_total",
		Help:      "Total number of cache hits",
	})
	misses := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "misses_total",
		Help:      "Total number of cache misses",
	})
	evictions := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "evictions_total",
		Help:      "Total number of cache evictions",
	})
	keys := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "keys_total",
		Help:      "Current number of keys in the cache",
	})
	reg.MustRegister(hits, misses, evictions, keys)
	return &Metrics{Hits: hits, Misses: misses, Evictions: evictions, Keys: keys}
}
