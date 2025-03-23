package id

import (
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// ValidatePage ...
func ValidatePage(page *paging.Page) {
	if page == nil {
		// unpaged
		return
	}

	switch page.Order() {
	// If the page order is ascending,
	// ensure that a minimum value is set.
	// This will be used as the cursor.
	case paging.OrderAscending:
		if page.Min.Value == "" {
			page.Min.Value = Lowest
		}

	// If the page order is descending,
	// ensure that a maximum value is set.
	// This will be used as the cursor.
	case paging.OrderDescending:
		if page.Max.Value == "" {
			page.Max.Value = Highest
		}
	}
}
