package cache

import (
	"sync"
	"time"
)

type (
	// CacheableEntity represents an entity that can be cached.
	// It must implement the Timestamp method to provide its last update time.
	CacheableEntity interface {
		Timestamp() time.Time
	}

	// Buffer is a thread-safe buffer for caching entities.
	// It stores a limited number of entities and removes the oldest ones when the limit is reached.
	Buffer[T CacheableEntity] struct {
		items []T
		// maxSize is the maximum number of items in the buffer.
		maxSize int
		// ttl is the time-to-live for items in the buffer.
		ttl time.Duration
		mu  sync.RWMutex
	}
)

// NewBuffer creates a new instance of Buffer for caching Price entities.
func NewBuffer[T CacheableEntity](ttl, expirationInternal time.Duration, maxSize int) *Buffer[T] {
	b := &Buffer[T]{
		items:   make([]T, 0),
		ttl:     ttl,
		maxSize: maxSize,
	}

	go b.trimExpired(expirationInternal)

	return b
}

// Add adds a new entity to the buffer.
// If the buffer is full, it removes the oldest entity.
// It is safe to call this method concurrently.
func (b *Buffer[T]) Add(update T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.items) == b.maxSize {
		// Remove oldest
		b.items = b.items[1:]
	}

	b.items = append(b.items, update)
}

// Last returns the last added entity in the buffer.
func (b *Buffer[T]) Last() T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.items) == 0 {
		var zero T
		return zero
	}

	return b.items[len(b.items)-1]
}

// Since retrieves all entities that were updated since the given time.
// It is safe to call this method concurrently.
func (b *Buffer[T]) Since(since time.Time) []T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []T

	for _, u := range b.items {
		if u.Timestamp().After(since) {
			result = append(result, u)
		}
	}

	return result
}

// Len returns the number of items in the buffer.
// It is safe to call this method concurrently.
func (b *Buffer[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.items)
}

func (b *Buffer[T]) trimExpired(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for range ticker.C {
		cutoff := time.Now().UTC().Add(-b.ttl)
		b.mu.Lock()

		var i int
		for ; i < len(b.items); i++ {
			if b.items[i].Timestamp().After(cutoff) {
				break
			}
		}

		b.items = b.items[i:] // trim old
		b.mu.Unlock()
	}
}
