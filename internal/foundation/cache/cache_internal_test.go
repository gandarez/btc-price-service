package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
	buffer := NewBuffer[mockEntity](2*time.Second, 10*time.Second, 5)

	require.NotNil(t, buffer)
	assert.Equal(t, 0, buffer.Len())
	assert.Equal(t, 5, buffer.maxSize)
	assert.Equal(t, 2*time.Second, buffer.ttl)
}

type mockEntity struct {
	UpdatedAt time.Time
}

func (m mockEntity) Timestamp() time.Time {
	return m.UpdatedAt
}
