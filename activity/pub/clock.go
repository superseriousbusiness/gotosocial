package pub

import (
	"time"
)

// Clock determines the time.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}
