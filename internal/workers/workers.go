/*
GoToSocial
Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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

	// Processor / federator worker pools.
	// ClientAPI runners.WorkerPool
	// Federator runners.WorkerPool

	// Media manager worker pools.
	Media runners.WorkerPool

	// prevent pass-by-value.
	_ nocopy
}

// Start will start all of the contained worker pools (and global scheduler).
func (w *Workers) Start() {
	// Get currently set GOMAXPROCS.
	maxprocs := runtime.GOMAXPROCS(0)

	tryUntil("starting scheduler", 5, func() bool {
		return w.Scheduler.Start(nil)
	})

	// tryUntil("starting client API workerpool", 5, func() bool {
	// 	return w.ClientAPI.Start(4*maxprocs, 400*maxprocs)
	// })

	// tryUntil("starting federator workerpool", 5, func() bool {
	// 	return w.Federator.Start(4*maxprocs, 400*maxprocs)
	// })

	tryUntil("starting media workerpool", 5, func() bool {
		return w.Media.Start(8*maxprocs, 80*maxprocs)
	})
}

// Stop will stop all of the contained worker pools (and global scheduler).
func (w *Workers) Stop() {
	tryUntil("stopping scheduler", 5, w.Scheduler.Stop)
	// tryUntil("stopping client API workerpool", 5, w.ClientAPI.Stop)
	// tryUntil("stopping federator workerpool", 5, w.Federator.Stop)
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
