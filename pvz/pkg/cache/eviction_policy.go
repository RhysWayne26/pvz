package cache

import "time"

type EvictionPolicy[K comparable, V any] interface {
	ShouldEvict(key K, item CachedItem[V]) bool
	OnAccess(key K)
	OnInsert(key K)
	OnDelete(key K)
	Name() string
}

type TTLPolicy[K comparable, V any] struct{}

func NewTTLPolicy[K comparable, V any]() *TTLPolicy[K, V] {
	return &TTLPolicy[K, V]{}
}

func (p *TTLPolicy[K, V]) ShouldEvict(_ K, item CachedItem[V]) bool {
	return !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt)
}
func (p *TTLPolicy[K, V]) OnAccess(_ K) {}
func (p *TTLPolicy[K, V]) OnInsert(_ K) {}
func (p *TTLPolicy[K, V]) OnDelete(_ K) {}
func (p *TTLPolicy[K, V]) Name() string { return "TTL with lazy expiration" }
