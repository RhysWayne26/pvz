package policies

import (
	"pvz-cli/pkg/cache/models"
)

// EvictionPolicy defines an interface for eviction strategies in a caching system to manage stored items efficiently.
type EvictionPolicy[K comparable, V any] interface {
	Evict(key K, item models.CachedItem[V]) (evictKey K, ok bool)
	OnAccess(key K)
	OnInsert(key K)
	OnDelete(key K)
	Name() string
}
