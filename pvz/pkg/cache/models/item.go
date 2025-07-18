package models

import "time"

// CachedItem represents a cached value and its expiration time.
type CachedItem[V any] struct {
	Value     V
	ExpiresAt time.Time
}
