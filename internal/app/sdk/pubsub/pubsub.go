package pubsub

import (
	"sync"

	"github.com/gandarez/btc-price-service/internal/foundation/cache"
)

type (
	// Subscriber represents a client that subscribes to updates.
	Subscriber struct {
		Ch   chan cache.CacheableEntity
		Done chan struct{} // signal to remove subscriber
	}

	// Broadcaster is responsible for managing subscribers and broadcasting updates.
	Broadcaster struct {
		subscribers map[*Subscriber]struct{}
		mu          sync.RWMutex
	}
)

// NewBroadcaster creates a new Broadcaster instance.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[*Subscriber]struct{}),
	}
}

// Subscribe adds a new subscriber to the broadcaster and returns a channel to receive updates.
func (b *Broadcaster) Subscribe() *Subscriber {
	sub := &Subscriber{
		Ch:   make(chan cache.CacheableEntity, 100), // buffered channel to avoid blocking
		Done: make(chan struct{}),
	}

	b.mu.Lock()
	b.subscribers[sub] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-sub.Done
		b.Unsubscribe(sub)
	}()

	return sub
}

// Unsubscribe removes a subscriber from the broadcaster.
func (b *Broadcaster) Unsubscribe(sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.subscribers, sub)
	close(sub.Ch)
}

// Broadcast sends an update to all subscribers.
func (b *Broadcaster) Broadcast(update cache.CacheableEntity) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub := range b.subscribers {
		select {
		case sub.Ch <- update:
		default: // skip slow clients
		}
	}
}
