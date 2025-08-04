package page_test

import (
	"testing"

	"github.com/gandarez/btc-price-service/internal/business/sdk/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageNew(t *testing.T) {
	p, err := page.New(1, 10)
	require.NoError(t, err)

	assert.Equal(t, p.Number(), 1)
	assert.Equal(t, p.RowsPerPage(), 10)
}

func TestPageNew_Err(t *testing.T) {
	tests := map[string]struct {
		number      int
		rowsPerPage int
		expected    string
	}{
		"page too small": {
			number:      0,
			rowsPerPage: 10,
			expected:    "page value too small, must be larger than 0",
		},
		"rows per page too small": {
			number:      1,
			rowsPerPage: 0,
			expected:    "rows value too small, must be larger than 0",
		},
		"rows per page too large": {
			number:      1,
			rowsPerPage: 101,
			expected:    "rows value too large, must be less than 100",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := page.New(test.number, test.rowsPerPage)

			assert.EqualError(t, err, test.expected)
			assert.Empty(t, p)
		})
	}
}

func TestPageString(t *testing.T) {
	p, err := page.New(3, 25)
	require.NoError(t, err)

	assert.Equal(t, "page: 3 rows: 25", p.String())
}
