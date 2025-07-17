package cache

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"hash/fnv"
	"math/bits"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

const baseShardsCount = 16

var _ Cache[string, any] = (*InMemoryShardedCache[string, any])(nil)

// CachedItem represents a cached value and its expiration time.
type CachedItem[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// InMemoryShardedCache is a high-performance, sharded in-memory key-value cache with TTL and eviction policies.
// It distributes keys across shards to minimize contention and supports concurrent read/write operations.
type InMemoryShardedCache[K comparable, V any] struct {
	shards         []shard[K, V]
	maxSize        int
	ttl            time.Duration
	stats          Stats
	evictionPolicy EvictionPolicy[K, V]
	group          singleflight.Group
	stopCh         chan struct{}
}

type shard[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]CachedItem[V]
}

// NewInMemoryShardedCache creates a new in-memory sharded cache with the specified number of shards.
func NewInMemoryShardedCache[K comparable, V any](shardsCount int) *InMemoryShardedCache[K, V] {
	if shardsCount <= 0 {
		shardsCount = baseShardsCount
	}

	if shardsCount&(shardsCount-1) != 0 {
		shardsCount = 1 << bits.Len(uint(shardsCount-1))
	}

	shards := make([]shard[K, V], shardsCount)
	for i := range shards {
		shards[i].items = make(map[K]CachedItem[V])
	}

	return &InMemoryShardedCache[K, V]{
		shards:         shards,
		stats:          Stats{},
		evictionPolicy: NewTTLPolicy[K, V](),
		stopCh:         make(chan struct{}),
	}
}

// Get retrieves the value associated with the given key from the cache and checks its validity based on the eviction policy.
func (c *InMemoryShardedCache[K, V]) Get(key K) (V, bool) {
	var zeroVal V
	s := c.getShard(key)
	s.mu.RLock()
	extracted, exists := s.items[key]
	if !exists {
		s.mu.RUnlock()
		atomic.AddInt64(&c.stats.Misses, 1)
		return zeroVal, false
	}

	if c.evictionPolicy.ShouldEvict(key, extracted) {
		s.mu.RUnlock()
		s.mu.Lock()
		if rechecked, ok := s.items[key]; ok && c.evictionPolicy.ShouldEvict(key, rechecked) {
			delete(s.items, key)
			c.evictionPolicy.OnDelete(key)
			atomic.AddInt64(&c.stats.Evictions, 1)
		}
		s.mu.Unlock()
		atomic.AddInt64(&c.stats.Misses, 1)
		return zeroVal, false
	} else {
		s.mu.RUnlock()
	}

	c.evictionPolicy.OnAccess(key)
	atomic.AddInt64(&c.stats.Hits, 1)
	return extracted.Value, true
}

// Set stores a value in the cache with the specified key, value, and time-to-live (ttl) duration.
func (c *InMemoryShardedCache[K, V]) Set(key K, value V, ttl time.Duration) {
	s := c.getShard(key)
	expiresAt := time.Now().Add(ttl)
	if ttl <= 0 {
		expiresAt = time.Time{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	newCachedItem := CachedItem[V]{
		Value:     value,
		ExpiresAt: expiresAt,
	}
	s.items[key] = newCachedItem
	c.evictionPolicy.OnInsert(key)
}

// GetOrSet retrieves the value associated with the given key or generates it using the factory if not present in the cache.
func (c *InMemoryShardedCache[K, V]) GetOrSet(key K, factory func() (V, error), ttl time.Duration) (V, error) {
	var zeroVal V
	result, err, _ := c.group.Do(fmt.Sprint(key), func() (interface{}, error) {
		if v, ok := c.Get(key); ok {
			return v, nil
		}
		v, err := factory()
		if err != nil {
			return nil, err
		}
		c.Set(key, v, ttl)
		return v, nil
	})
	if err != nil {
		return zeroVal, err
	}

	return result.(V), nil
}

// Invalidate removes the specified key from the cache and triggers the eviction policy if the key exists.
func (c *InMemoryShardedCache[K, V]) Invalidate(key K) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[key]; ok {
		delete(s.items, key)
		c.evictionPolicy.OnDelete(key)
		atomic.AddInt64(&c.stats.Evictions, 1)
	}
}

// InvalidateAll clears all items from all shards in the cache, applying the eviction policy for deleted keys.
func (c *InMemoryShardedCache[K, V]) InvalidateAll() {
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for k := range s.items {
			delete(s.items, k)
			c.evictionPolicy.OnDelete(k)
			atomic.AddInt64(&c.stats.Evictions, 1)
		}
		s.mu.Unlock()
	}
}

// InvalidatePattern invalidates all cache entries whose keys match the provided regex pattern.
func (c *InMemoryShardedCache[K, V]) InvalidatePattern(pattern string) {
	rgx := regexp.MustCompile(pattern)
	c.InvalidateFunc(func(key K) bool {
		return rgx.MatchString(fmt.Sprint(key))
	})
}

// InvalidateFunc removes items from the cache based on the provided function `fn` which determines keys to invalidate.
func (c *InMemoryShardedCache[K, V]) InvalidateFunc(fn func(key K) bool) {
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for key := range s.items {
			if fn(key) {
				delete(s.items, key)
				c.evictionPolicy.OnDelete(key)
				atomic.AddInt64(&c.stats.Evictions, 1)
			}
		}
		s.mu.Unlock()
	}
}

// Has checks if the given key exists in the cache. Returns true if the key is present, otherwise false.
func (c *InMemoryShardedCache[K, V]) Has(key K) bool {
	_, ok := c.Get(key)
	return ok
}

// TTL returns the remaining time-to-live (TTL) for the specified key in the cache. If the key does not exist, it returns -1. If the key has no expiration, it returns 0.
func (c *InMemoryShardedCache[K, V]) TTL(key K) time.Duration {
	s := c.getShard(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[key]
	if !exists {
		return -1
	}
	if item.ExpiresAt.IsZero() {
		return 0
	}
	remainingTTL := time.Until(item.ExpiresAt)
	if remainingTTL < 0 {
		return 0
	}
	return remainingTTL
}

// Keys returns a slice of all the keys currently stored in the cache across all shards, excluding expired or evicted items.
func (c *InMemoryShardedCache[K, V]) Keys() []K {
	keys := make([]K, 0)
	now := time.Now()
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.RLock()
		for key, item := range s.items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				continue
			}
			if c.evictionPolicy != nil && c.evictionPolicy.ShouldEvict(key, item) {
				continue
			}
			keys = append(keys, key)
		}
		s.mu.RUnlock()
	}
	return keys
}

// Items retrieves all the items currently stored in the cache, excluding expired items or those evicted by the policy.
func (c *InMemoryShardedCache[K, V]) Items() map[K]V {
	values := make(map[K]V)
	now := time.Now()
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.RLock()
		for key, item := range s.items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				continue
			}
			if c.evictionPolicy != nil && c.evictionPolicy.ShouldEvict(key, item) {
				continue
			}
			values[key] = item.Value
		}
		s.mu.RUnlock()
	}
	return values
}

// Size calculates and returns the number of items currently stored in the cache across all shards. It skips items marked for eviction by the eviction policy.
func (c *InMemoryShardedCache[K, V]) Size() int { // NOT O(1) OPERATION A THIS POINT, BE CAREFUL (MESSAGE TO MYSELF)
	size := 0
	for _, s := range c.shards {
		s.mu.RLock()
		for k, v := range s.items {
			if !c.evictionPolicy.ShouldEvict(k, v) {
				size++
			}
		}

		s.mu.RUnlock()
	}
	return size
}

// MaxSize returns the maximum size limit for the cache.
func (c *InMemoryShardedCache[K, V]) MaxSize() int {
	return c.maxSize
}

// SizeRatio returns a string representing the current size of the cache relative to its maximum size.
func (c *InMemoryShardedCache[K, V]) SizeRatio() string {
	current := c.Size()
	if c.maxSize <= 0 {
		return fmt.Sprintf("%d/∞", current)
	}
	return fmt.Sprintf("%d/%d", current, c.maxSize)
}

// Stats returns the current cache statistics.
func (c *InMemoryShardedCache[K, V]) Stats() Stats {
	return Stats{
		Hits:        atomic.LoadInt64(&c.stats.Hits),
		Misses:      atomic.LoadInt64(&c.stats.Misses),
		Evictions:   atomic.LoadInt64(&c.stats.Evictions),
		KeysTotal:   c.Size(),
		MemoryUsage: 456, //hardcode, add calculation logic
	}
}

// PurgeExpired removes all expired items from the in-memory sharded cache.
func (c *InMemoryShardedCache[K, V]) PurgeExpired() {
	now := time.Now()
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for key, item := range s.items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				delete(s.items, key)
				c.evictionPolicy.OnDelete(key)
				atomic.AddInt64(&c.stats.Evictions, 1)
			}
		}
		s.mu.Unlock()
	}
}

// UpdateTTL updates the time-to-live of the specified key in the cache and returns true if the key exists and was updated.
func (c *InMemoryShardedCache[K, V]) UpdateTTL(key K, ttl time.Duration) bool {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	item, exists := s.items[key]
	if !exists {
		return false
	}

	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		delete(s.items, key)
		c.evictionPolicy.OnDelete(key)
		atomic.AddInt64(&c.stats.Evictions, 1)
		return false
	}

	if ttl <= 0 {
		item.ExpiresAt = time.Time{}
	} else {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	s.items[key] = item
	return true
}

// ResetStats resets the statistics of the cache to zero atomically.
func (c *InMemoryShardedCache[K, V]) ResetStats() {
	atomic.StoreInt64(&c.stats.Hits, 0)
	atomic.StoreInt64(&c.stats.Misses, 0)
	atomic.StoreInt64(&c.stats.Evictions, 0)
}

// EvictionPolicy returns the name of the current eviction policy being used by the cache.
func (c *InMemoryShardedCache[K, V]) EvictionPolicy() string {
	return c.evictionPolicy.Name()
}

// Close releases all resources associated with the cache and clears all stored items.
func (c *InMemoryShardedCache[K, V]) Close() error {
	close(c.stopCh)
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		for key := range s.items {
			delete(s.items, key)
		}
		s.mu.Unlock()
	}
	return nil
}

func (c *InMemoryShardedCache[K, V]) getShard(key K) *shard[K, V] {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprint(key)))
	idx := h.Sum64() & (uint64(len(c.shards)) - 1)
	return &c.shards[idx]
}

func (c *InMemoryShardedCache[K, V]) startBackgroundPurge(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.PurgeExpired()
		case <-c.stopCh:
			return
		}
	}
}
