package ttl

import (
	"time"

	"codeberg.org/gruf/go-sched"
)

// scheduler is the global cache runtime
// scheduler for handling cache evictions.
var scheduler sched.Scheduler

// schedule will given sweep  routine to the global scheduler, and start global scheduler.
func schedule(sweep func(time.Time), freq time.Duration) func() {
	if !scheduler.Running() {
		// ensure sched running
		_ = scheduler.Start(nil)
	}
	return scheduler.Schedule(sched.NewJob(sweep).Every(freq))
}
