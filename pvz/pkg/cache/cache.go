package cache

import "time"

// Cache represents a generic interface for a key-value cache with expiration and various management features.
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
	GetOrSet(key K, factory func() (V, error), ttl time.Duration) (V, error)

	Invalidate(key K)
	InvalidateAll()
	InvalidatePattern(pattern string)
	InvalidateFunc(fn func(key K) bool)

	Has(key K) bool
	TTL(key K) time.Duration

	Keys() []K
	Items() map[K]V
	Size() int
	MaxSize() int
	SizeRatio() string
	Stats() Stats

	PurgeExpired()
	UpdateTTL(key K, ttl time.Duration) bool
	ResetStats()
	EvictionPolicy() string

	Close() error
}

// Stats represents cache statistics information such as hits, misses, evictions, total keys, and memory usage.
type Stats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	KeysTotal   int
	MemoryUsage int64
}
