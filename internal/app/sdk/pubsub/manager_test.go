package pubsub_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gandarez/btc-price-service/internal/app/sdk/pubsub"
)

func TestManager_DataRace(t *testing.T) {
	m := pubsub.NewManager(2)

	var mu sync.Mutex

	// Start broadcasting updates
	go func() {
		for {
			m.Broadcast(mockEntity{UpdatedAt: time.Now().UTC()})

			time.Sleep(50 * time.Millisecond)
		}
	}()

	subscribers := make([]*pubsub.Subscriber, 0, 100)

	var wg sync.WaitGroup
	wg.Add(100)

	go func() {
		for range 100 {
			go func() {
				defer wg.Done()

				sub := m.Subscribe(t.Context())
				require.NotNil(t, sub)

				mu.Lock()

				subscribers = append(subscribers, sub)

				mu.Unlock()

				time.Sleep(time.Duration(rand.Intn(50-10)+10) * time.Millisecond) // Simulate some delay
			}()
		}
	}()

	wg.Wait()

	// Randomly unsubscribe some subscribers
	for _, sub := range subscribers {
		go func() {
			time.Sleep(time.Duration(rand.Intn(20-1)+1) * time.Millisecond) // Simulate some delay

			m.Unsubscribe(t.Context(), sub)
		}()
	}

	assert.Eventually(t, func() bool { return m.SubscribersCount() == 0 }, 4*time.Second, 100*time.Millisecond)
}

func TestManager_Subscribe(t *testing.T) {
	m := pubsub.NewManager(3)

	sub := m.Subscribe(t.Context())
	require.NotNil(t, sub)
	assert.NotEmpty(t, sub.BroadcasterID())

	broadcasterID := sub.BroadcasterID()

	broadcaster, exists := m.GetBroadcaster(broadcasterID)
	require.True(t, exists)
	require.NotNil(t, broadcaster)
	assert.Equal(t, 1, broadcaster.Len())

	assert.Equal(t, 1, m.PoolLen())
}

func TestManager_Subscribe_MaxPeersPerBroadcaster(t *testing.T) {
	m := pubsub.NewManager(2)

	sub1 := m.Subscribe(t.Context())
	require.NotNil(t, sub1)
	assert.NotEmpty(t, sub1.BroadcasterID())

	sub2 := m.Subscribe(t.Context())
	require.NotNil(t, sub2)
	assert.NotEmpty(t, sub2.BroadcasterID())

	sub3 := m.Subscribe(t.Context())
	require.NotNil(t, sub3)
	assert.NotEmpty(t, sub3.BroadcasterID())

	assert.Equal(t, sub1.BroadcasterID(), sub2.BroadcasterID())
	assert.NotEqual(t, sub1.BroadcasterID(), sub3.BroadcasterID())
	assert.NotEqual(t, sub2.BroadcasterID(), sub3.BroadcasterID())

	broadcasterID1 := sub1.BroadcasterID()
	broadcaster, exists := m.GetBroadcaster(broadcasterID1)
	require.True(t, exists)
	require.NotNil(t, broadcaster)
	assert.Equal(t, 2, broadcaster.Len())

	broadcasterID2 := sub3.BroadcasterID()
	broadcaster2, exists := m.GetBroadcaster(broadcasterID2)
	require.True(t, exists)
	require.NotNil(t, broadcaster2)
	assert.Equal(t, 1, broadcaster2.Len())

	assert.Equal(t, 2, m.PoolLen())
}

func TestManager_Unsubscribe(t *testing.T) {
	m := pubsub.NewManager(1)

	sub := m.Subscribe(t.Context())
	require.NotNil(t, sub)

	m.Unsubscribe(t.Context(), sub)

	assert.Zero(t, m.PoolLen())

	broadcasterID := sub.BroadcasterID()

	broadcaster, exists := m.GetBroadcaster(broadcasterID)
	require.False(t, exists)
	assert.Nil(t, broadcaster)
}

func TestManager_Broadcast(t *testing.T) {
	m := pubsub.NewManager(1)

	sub1 := m.Subscribe(t.Context())
	require.NotNil(t, sub1)

	sub2 := m.Subscribe(t.Context())
	require.NotNil(t, sub2)

	now := time.Now().UTC()

	m.Broadcast(mockEntity{UpdatedAt: now})

	select {
	case received := <-sub1.Ch:
		assert.Equal(t, now, received.Timestamp())
	case received := <-sub2.Ch:
		assert.Equal(t, now, received.Timestamp())
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}

func TestManager_SubscriberCounts(t *testing.T) {
	m := pubsub.NewManager(2)

	sub1 := m.Subscribe(t.Context())
	require.NotNil(t, sub1)

	sub2 := m.Subscribe(t.Context())
	require.NotNil(t, sub2)

	sub3 := m.Subscribe(t.Context())
	require.NotNil(t, sub3)

	assert.Equal(t, 3, m.SubscribersCount())
}

func TestManager_SendOne(t *testing.T) {
	m := pubsub.NewManager(1)

	sub := m.Subscribe(t.Context())
	require.NotNil(t, sub)

	now := time.Now().UTC()

	m.SendOne(sub, mockEntity{UpdatedAt: now})

	select {
	case received := <-sub.Ch:
		assert.Equal(t, now, received.Timestamp())
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}
