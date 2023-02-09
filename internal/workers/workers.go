package workers

import (
	"log"
	"runtime"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-sched"
)

type Workers struct {
	// Main task scheduler instance.
	Scheduler sched.Scheduler
	SchedJobs []func(*sched.Scheduler)

	// Processor / federator worker pools.
	ClientAPI runners.WorkerPool
	Federator runners.WorkerPool

	// Media manager worker pools.
	Media runners.WorkerPool

	// prevent pass-by-value.
	_ nocopy
}

func (w *Workers) Start() {
	// Get currently set GOMAXPROCS.
	maxprocs := runtime.GOMAXPROCS(0)

	tryUntil("starting scheduler", 5, func() bool {
		return w.Scheduler.Start(nil)
	})

	tryUntil("starting client API workerpool", 5, func() bool {
		return w.ClientAPI.Start(4*maxprocs, 400*maxprocs)
	})

	tryUntil("starting federator workerpool", 5, func() bool {
		return w.Federator.Start(4*maxprocs, 400*maxprocs)
	})

	tryUntil("starting media workerpool", 5, func() bool {
		return w.Media.Start(8*maxprocs, 80*maxprocs)
	})
}

func (w *Workers) Stop() {
	tryUntil("stopping scheduler", 5, w.Scheduler.Stop)
	tryUntil("stopping client API workerpool", 5, w.ClientAPI.Stop)
	tryUntil("stopping federator workerpool", 5, w.Federator.Stop)
	tryUntil("stopping media workerpool", 5, w.Media.Stop)
}

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}

// tryUntil will attempt to call 'do' for 'count' attempts, before panicking with 'msg'.
func tryUntil(msg string, count int, do func() bool) {
	for i := 0; i < count; i++ {
		if do() {
			return
		}
	}
	log.Panicf("failed %s after %d tries", msg, count)
}
