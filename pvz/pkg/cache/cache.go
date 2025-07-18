package cache

import (
	"pvz-cli/pkg/cache/models"
	"time"
)

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
	Stats() models.Stats

	PurgeExpired()
	UpdateTTL(key K, ttl time.Duration) bool
	ResetStats()
	EvictionPolicy() string

	Close() error
}
