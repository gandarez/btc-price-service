package cache_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gandarez/btc-price-service/internal/foundation/cache"
)

func TestBuffer_TTL(t *testing.T) {
	buffer := cache.NewBuffer[mockEntity](20*time.Millisecond, 10*time.Millisecond, 5)
	require.NotNil(t, buffer)

	oldTime := time.Now().UTC().Add(-30 * time.Millisecond)
	buffer.Add(mockEntity{UpdatedAt: oldTime})

	require.Equal(t, 1, buffer.Len())

	assert.Eventually(t, func() bool {
		return buffer.Len() == 0
	}, 50*time.Millisecond, 10*time.Millisecond)
}

func TestBuffer_MaxSize(t *testing.T) {
	buffer := cache.NewBuffer[mockEntity](2*time.Second, 10*time.Second, 5)
	require.NotNil(t, buffer)

	for range 10 {
		buffer.Add(mockEntity{UpdatedAt: time.Now().UTC()})
	}

	assert.Equal(t, 5, buffer.Len())
}

func TestBuffer_Since(t *testing.T) {
	buffer := cache.NewBuffer[mockEntity](10*time.Second, 20*time.Second, 5)
	require.NotNil(t, buffer)

	now := time.Now().UTC()
	for i := range 5 {
		buffer.Add(mockEntity{UpdatedAt: now.Add(time.Duration(i) * time.Second)})
	}

	since := now.Add(1 * time.Second)
	updates := buffer.Since(since)

	require.Len(t, updates, 3)

	for _, update := range updates {
		assert.True(t, update.UpdatedAt.After(since))
	}
}

type mockEntity struct {
	UpdatedAt time.Time
}

func (m mockEntity) Timestamp() time.Time {
	return m.UpdatedAt
}
