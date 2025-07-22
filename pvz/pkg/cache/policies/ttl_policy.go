package policies

import (
	"pvz-cli/pkg/cache/models"
	"time"
)

var _ EvictionPolicy[any, any] = (*TTLPolicy[any, any])(nil)

// TTLPolicy defines a time-to-live policies policy for cache entries with lazy expiration logic.
type TTLPolicy[K comparable, V any] struct{}

// NewTTLPolicy returns a new instance of TTLPolicy with lazy expiration for cache entries.
func NewTTLPolicy[K comparable, V any]() *TTLPolicy[K, V] {
	return &TTLPolicy[K, V]{}
}

func (p *TTLPolicy[K, V]) Evict(key K, item models.CachedItem[V]) (K, bool) {
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		return key, true
	}
	var zero K
	return zero, false
}

// OnAccess is a no-op method for handling access events on a cached item in the TTL policy.
func (p *TTLPolicy[K, V]) OnAccess(_ K) {}

// OnInsert is a no-op method invoked when a new item is inserted into the cache.
func (p *TTLPolicy[K, V]) OnInsert(_ K) {}

// OnDelete is a no-op method invoked when an item is deleted from the cache.
func (p *TTLPolicy[K, V]) OnDelete(_ K) {}

// Name returns the name of the TTL policies policy as a string.
func (p *TTLPolicy[K, V]) Name() string { return "TTL with lazy expiration" }
