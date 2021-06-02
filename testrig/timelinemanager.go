package testrig

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
)

func NewTestTimelineManager(db db.DB) timeline.Manager {
	return timeline.NewManager(db, NewTestTypeConverter(db), NewTestConfig(), NewTestLog())
}
