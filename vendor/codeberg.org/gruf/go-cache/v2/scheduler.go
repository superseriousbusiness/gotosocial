package cache

import (
	"time"

	"codeberg.org/gruf/go-sched"
)

// scheduler is the global cache runtime scheduler
// for handling regular cache evictions.
var scheduler = sched.NewScheduler(5)

// schedule will given sweep  routine to the global scheduler, and start global scheduler.
func schedule(sweep func(time.Time), freq time.Duration) func() {
	go scheduler.Start() // does nothing if already running
	return scheduler.Schedule(sched.NewJob(sweep).Every(freq))
}
