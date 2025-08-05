package pubsub

import (
	"context"
	"sync"

	"github.com/gandarez/btc-price-service/internal/foundation/cache"
	"github.com/gandarez/btc-price-service/internal/foundation/log"
)

// Manager manages multiple broadcasters and their subscribers.
type Manager struct {
	maxPeersPerBroadcaster int
	pool                   map[string]*Broadcaster
	mu                     sync.RWMutex
}

// NewManager creates a new Manager instance.
func NewManager(maxPeersPerBroadcaster int) *Manager {
	return &Manager{
		maxPeersPerBroadcaster: maxPeersPerBroadcaster,
		pool:                   make(map[string]*Broadcaster),
	}
}

// Subscribe adds a new subscriber to the most appropriate broadcaster.
func (m *Manager) Subscribe(ctx context.Context) *Subscriber {
	logger := log.Extract(ctx)

	m.mu.RLock()

	for _, b := range m.pool {
		if len(b.subscribers) < m.maxPeersPerBroadcaster {
			logger.Infof("reusing broadcaster %s with %d subscribers", b.id, len(b.subscribers))

			sub := b.Subscribe()

			m.mu.RUnlock()

			return sub
		}
	}

	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	broadcaster := NewBroadcaster()
	m.pool[broadcaster.id] = broadcaster

	logger.Infof("created new broadcaster %s", broadcaster.id)

	sub := broadcaster.Subscribe()

	m.redistributeSubscribers(ctx)

	return sub
}

// Unsubscribe removes a subscriber from its broadcaster and redistributes subscribers if needed.
func (m *Manager) Unsubscribe(ctx context.Context, sub *Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := log.Extract(ctx)
	logger.Infof("unsubscribing from broadcaster %s", sub.broadcasterID)

	m.pool[sub.broadcasterID].Unsubscribe(sub)

	// Clean up broadcaster if no subscribers left
	if len(m.pool[sub.broadcasterID].subscribers) == 0 {
		logger.Infof("removing broadcaster %s with no subscribers", sub.broadcasterID)

		delete(m.pool, sub.broadcasterID)
	}

	m.redistributeSubscribers(ctx)
}

// Broadcast sends an update to all subscribers across all broadcasters.
func (m *Manager) Broadcast(update cache.CacheableEntity) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup

	for _, b := range m.pool {
		wg.Add(1)

		go func(b *Broadcaster) {
			defer wg.Done()

			b.Broadcast(update)
		}(b)
	}

	wg.Wait()
}

// SendOne sends an update to a specific subscriber.
func (m *Manager) SendOne(sub *Subscriber, update cache.CacheableEntity) {
	if b, ok := m.pool[sub.broadcasterID]; ok {
		b.SendOne(sub, update)
	}
}

// PoolLen returns the number of broadcasters in the pool.
func (m *Manager) PoolLen() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.pool)
}

// SubscribersCount returns the total number of subscribers across all broadcasters.
func (m *Manager) SubscribersCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int
	for _, b := range m.pool {
		count += b.Len()
	}

	return count
}

// GetBroadcaster retrieves a broadcaster by its ID.
// Returns the broadcaster and a boolean indicating if it exists.
func (m *Manager) GetBroadcaster(id string) (*Broadcaster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	broadcaster, exists := m.pool[id]

	return broadcaster, exists
}

// redistributeSubscribers moves subscribers from most populated to least populated broadcasters to balance the load.
func (m *Manager) redistributeSubscribers(ctx context.Context) {
	if len(m.pool) <= 1 {
		return // No need to rebalance with 0 or 1 broadcaster
	}

	logger := log.Extract(ctx)

	var (
		maxBroadcaster   *Broadcaster
		minBroadcaster   *Broadcaster
		minBroadcasterID string
	)

	maxCount := -1
	minCount := m.maxPeersPerBroadcaster + 1

	for id, b := range m.pool {
		// Find broadcaster with most subscribers
		if len(b.subscribers) > maxCount {
			maxCount = len(b.subscribers)
			maxBroadcaster = b
		}

		// Find broadcaster with least subscribers
		if len(b.subscribers) < minCount {
			minCount = len(b.subscribers)
			minBroadcaster = b
			minBroadcasterID = id
		}
	}

	// If difference is <= 1, we're balanced
	if maxCount-minCount <= 1 {
		logger.Infof("broadcasters balanced: max %d, min %d", maxCount, minCount)

		return
	}

	if minBroadcaster == nil || maxBroadcaster == nil {
		logger.Warnf("unable to find min or max broadcaster for redistribution")
		return
	}

	// Move one subscriber from max to min
	var subscriberToMove *Subscriber
	for sub := range maxBroadcaster.subscribers {
		subscriberToMove = sub
		break // get first subscriber
	}

	// Remove from max broadcaster
	delete(maxBroadcaster.subscribers, subscriberToMove)

	// Add to min broadcaster
	subscriberToMove.broadcasterID = minBroadcasterID
	minBroadcaster.subscribers[subscriberToMove] = struct{}{}

	logger.Infof("moved subscriber to broadcaster %s, new counts: max %d, min %d",
		minBroadcasterID,
		len(maxBroadcaster.subscribers),
		len(minBroadcaster.subscribers),
	)
}
