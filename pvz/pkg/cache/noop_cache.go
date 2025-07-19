package cache

import (
	"pvz-cli/pkg/cache/models"
	"time"
)

var _ Cache[string, any] = (*NoopCache)(nil)

// NoopCache is a no-operation cache implementation that performs no actual caching but satisfies the Cache interface.
type NoopCache struct{}

// NewNoopCache creates and returns a new instance of NoopCache, which satisfies the Cache interface but performs no caching.
func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

// Get retrieves a value from the cache by the specified key. Always returns nil and false in this implementation.
func (c *NoopCache) Get(key string) (any, bool) {
	return nil, false
}

// Set stores a value in the cache with the specified key and time-to-live. Performs no action in this implementation.
func (c *NoopCache) Set(key string, value any, ttl time.Duration) {}

// GetOrSet retrieves a value for the given key or generates it using the factory function. No value is cached.
func (c *NoopCache) GetOrSet(key string, factory func() (any, error), ttl time.Duration) (any, error) {
	return factory()
}

// Invalidate removes the specified key from the cache. Performs no action in this implementation.
func (c *NoopCache) Invalidate(key string) {}

// InvalidateAll clears all entries from the cache. Performs no action in this no-operation cache implementation.
func (c *NoopCache) InvalidateAll() {}

// InvalidatePattern attempts to remove cache entries matching the specified pattern. Performs no action in this implementation.
func (c *NoopCache) InvalidatePattern(pattern string) {}

// InvalidateFunc removes all cache entries for which the provided function returns true. Performs no action in this implementation.
func (c *NoopCache) InvalidateFunc(fn func(key string) bool) {}

// Has checks if the given key exists in the cache. Always returns false in this no-operation cache implementation.
func (c *NoopCache) Has(key string) bool {
	return false
}

// TTL returns the time-to-live of a specified key. Always returns -1 as this implementation does not support TTL.
func (c *NoopCache) TTL(key string) time.Duration {
	return -1
}

// Keys returns a slice of all the keys currently stored in the cache. Always returns an empty slice in this implementation.
func (c *NoopCache) Keys() []string {
	return nil
}

// Items returns a map of all items currently stored in the cache. Always returns an empty map in this no-operation implementation.
func (c *NoopCache) Items() map[string]any {
	return map[string]any{}
}

// Size returns the number of items currently stored in the cache. Always returns 0 in this no-operation implementation.
func (c *NoopCache) Size() int {
	return 0
}

// Stats provides the current cache statistics, including hits, misses, evictions, and total keys. Always returns default values.
func (c *NoopCache) Stats() models.Stats {
	return models.Stats{}
}

// PurgeExpired performs no action in this no-operation cache implementation. It satisfies the Cache interface method.
func (c *NoopCache) PurgeExpired() {}

// UpdateTTL attempts to update the TTL for the given key. Always returns false in this no-operation cache implementation.
func (c *NoopCache) UpdateTTL(key string, ttl time.Duration) bool {
	return false
}

// ResetStats resets all internal cache statistics such as hits, misses, and evictions in this no-operation implementation.
func (c *NoopCache) ResetStats() {}

// EvictionPolicy returns the eviction policy of the cache, which is always "noop" for the no-operation cache implementation.
func (c *NoopCache) EvictionPolicy() string {
	return "noop"
}

// Close performs any necessary cleanup for the cache. In this no-operation implementation, it simply returns nil.
func (c *NoopCache) Close() error {
	return nil
}
