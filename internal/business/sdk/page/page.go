// Package page provides support for query paging.
package page

import (
	"fmt"
)

// Page represents the requested page and rows per page.
type Page struct {
	number int
	rows   int
}

// New creates a new Page instance and validates the values are in reason.
func New(number, rowsPerPage int) (Page, error) {
	if number <= 0 {
		return Page{}, fmt.Errorf("page value too small, must be larger than 0")
	}

	if rowsPerPage <= 0 {
		return Page{}, fmt.Errorf("rows value too small, must be larger than 0")
	}

	if rowsPerPage > 100 {
		return Page{}, fmt.Errorf("rows value too large, must be less than 100")
	}

	p := Page{
		number: number,
		rows:   rowsPerPage,
	}

	return p, nil
}

// String implements the stringer interface.
func (p Page) String() string {
	return fmt.Sprintf("page: %d rows: %d", p.number, p.rows)
}

// Number returns the page number.
func (p Page) Number() int {
	return p.number
}

// RowsPerPage returns the rows per page.
func (p Page) RowsPerPage() int {
	return p.rows
}
