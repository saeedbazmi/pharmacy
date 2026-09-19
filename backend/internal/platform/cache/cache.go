package cache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	val T
	exp time.Time
}

// TTL is a small process-local cache. Fine for a single API node in phase 1.
type TTL[T any] struct {
	mu      sync.Mutex
	items   map[string]entry[T]
	ttl     time.Duration
	maxSize int
}

func NewTTL[T any](ttl time.Duration, maxSize int) *TTL[T] {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	if maxSize < 1 {
		maxSize = 512
	}
	return &TTL[T]{items: map[string]entry[T]{}, ttl: ttl, maxSize: maxSize}
}

func (c *TTL[T]) Get(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero T
	e, ok := c.items[key]
	if !ok {
		return zero, false
	}
	if time.Now().After(e.exp) {
		delete(c.items, key)
		return zero, false
	}
	return e.val, true
}

func (c *TTL[T]) Set(key string, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.maxSize {
		c.items = map[string]entry[T]{}
	}
	c.items[key] = entry[T]{val: val, exp: time.Now().Add(c.ttl)}
}

func (c *TTL[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}
