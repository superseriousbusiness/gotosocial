package cache

import (
	"time"
)

// reaper is the global cache eviction routine
func reaper() {
	// Reaper ticker, tick-tick!
	tick := time.Tick(
		clockPrecision * 100,
	)

	for {
		// Rest now little CPU,
		// save your cycles...
		<-tick

		// Attempt to sweep caches
		globalMutex.Lock()
		for _, cache := range caches {
			cache.sweep()
		}
		globalMutex.Unlock()
	}
}
