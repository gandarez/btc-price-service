package pubsub_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gandarez/btc-price-service/internal/app/sdk/pubsub"
)

func TestBroadcaster(t *testing.T) {
	b := pubsub.NewBroadcaster()

	sub := b.Subscribe()
	defer close(sub.Done)

	now := time.Now().UTC()

	b.Broadcast(mockEntity{UpdatedAt: now})

	// Test receiving the message
	select {
	case received := <-sub.Ch:
		assert.Equal(t, now, received.Timestamp())
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	b := pubsub.NewBroadcaster()

	sub := b.Subscribe()

	b.Broadcast(mockEntity{UpdatedAt: time.Now().UTC()})

	// Test receiving the message
	select {
	case <-sub.Ch:
		b.Unsubscribe(sub)
		assert.Empty(t, sub.Ch)
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for message")
	}
}

type mockEntity struct {
	UpdatedAt time.Time
}

func (m mockEntity) Timestamp() time.Time {
	return m.UpdatedAt
}
