package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
)

// NewTestTimelineManager retuts a new timeline.Manager, suitable for testing, using the given db.
func NewTestTimelineManager(db db.DB) timeline.Manager {
	return timeline.NewManager(db, NewTestTypeConverter(db), NewTestConfig())
}
