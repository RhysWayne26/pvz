package policies

import (
	"container/list"
	"fmt"
	"pvz-cli/pkg/cache/models"
	"sync"
)

var _ EvictionPolicy[any, any] = (*LRUPolicy[any, any])(nil)

// LRUPolicy is a thread-safe implementation of the Least Recently Used caching policy to manage key-value pairs.
type LRUPolicy[K comparable, V any] struct {
	mu    sync.Mutex
	cap   int
	dll   *list.List
	nodes map[K]*list.Element
}

// NewLRUPolicy initializes and returns a new instance of LRUPolicy with the specified capacity.
func NewLRUPolicy[K comparable, V any](cap int) *LRUPolicy[K, V] {
	return &LRUPolicy[K, V]{
		cap:   cap,
		dll:   list.New(),
		nodes: make(map[K]*list.Element, cap),
	}
}

// Evict removes the least recently used item from the cache if the capacity is exceeded and returns its key and success status.
func (p *LRUPolicy[K, V]) Evict(_ K, _ models.CachedItem[V]) (K, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.dll.Len() <= p.cap {
		var zero K
		return zero, false
	}
	tail := p.dll.Back()
	if tail == nil {
		var zero K
		return zero, false
	}
	item := tail.Value.(lruCacheNode[K])
	return item.key, true
}

// OnAccess moves the accessed key to the front of the cache to mark it as recently used, ensuring thread safety.
func (p *LRUPolicy[K, V]) OnAccess(key K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if element, ok := p.nodes[key]; ok {
		p.dll.MoveToFront(element)
	}
}

// OnInsert adds a key to the LRU cache or moves it to the front if it already exists, ensuring thread safety.
func (p *LRUPolicy[K, V]) OnInsert(key K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if element, exists := p.nodes[key]; exists {
		p.dll.MoveToFront(element)
		return
	}
	element := p.dll.PushFront(lruCacheNode[K]{key: key})
	p.nodes[key] = element
}

// OnDelete removes a key from the LRU cache, ensuring thread safety and updating the linked list and map accordingly.
func (p *LRUPolicy[K, V]) OnDelete(key K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if elem, exists := p.nodes[key]; exists {
		p.dll.Remove(elem)
		delete(p.nodes, key)
	}
}

// Name returns a string representing the name and the capacity of the LRU policy.
func (p *LRUPolicy[K, V]) Name() string {
	return fmt.Sprintf("LRU(cap=%d)", p.cap)
}

type lruCacheNode[K comparable] struct {
	key K
}
