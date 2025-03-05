// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package scheduler

import (
	"context"
	"sync"
	"time"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-sched"
)

// Scheduler wraps an underlying scheduler to provide
// task tracking by unique string identifiers, so jobs
// may be cancelled with only an identifier.
type Scheduler struct {
	sch sched.Scheduler
	ts  map[string]*task
	mu  sync.Mutex
}

// Start attempts to start the scheduler. Returns false if already running.
func (sch *Scheduler) Start() bool {
	if sch.sch.Start(nil) {
		sch.ts = make(map[string]*task)
		return true
	}
	return false
}

// Started returns true if the scheduler has been started.
func (sch *Scheduler) Started() bool {
	return sch.ts != nil
}

// Stop attempts to stop scheduler, cancelling
// all running tasks. Returns false if not running.
func (sch *Scheduler) Stop() bool {
	if sch.sch.Stop() {
		sch.ts = nil
		return true
	}
	return false
}

// AddOnce schedules the given task to run at time, registered under the given ID. Returns false if task already exists for id.
func (sch *Scheduler) AddOnce(id string, start time.Time, fn func(context.Context, time.Time)) bool {
	return sch.schedule(id, fn, (*sched.Once)(&start))
}

// AddRecurring schedules the given task to return at given period, starting at given time, registered under given id. Returns false if task already exists for id.
func (sch *Scheduler) AddRecurring(id string, start time.Time, freq time.Duration, fn func(context.Context, time.Time)) bool {
	return sch.schedule(id, fn, &sched.PeriodicAt{Once: sched.Once(start), Period: sched.Periodic(freq)})
}

// Cancel attempts to cancel a scheduled task with id, returns false if no task found.
func (sch *Scheduler) Cancel(id string) bool {
	// Attempt to acquire and
	// delete task with iD.
	sch.mu.Lock()
	task, ok := sch.ts[id]
	delete(sch.ts, id)
	sch.mu.Unlock()

	if !ok {
		// none found.
		return false
	}

	// Cancel the queued
	// job from Scheduler.
	task.cncl()
	return true
}

func (sch *Scheduler) schedule(id string, fn func(context.Context, time.Time), t sched.Timing) bool {
	if fn == nil {
		panic("nil function")
	}

	// Perform within lock.
	sch.mu.Lock()
	defer sch.mu.Unlock()

	if _, ok := sch.ts[id]; ok {
		// existing task already
		// exists under this ID.
		return false
	}

	// Extract current sched context.
	doneCh := sch.sch.Done()
	ctx := runners.CancelCtx(doneCh)

	// Create a new job to hold task function with
	// timing, passing in the current sched context.
	job := sched.NewJob(func(now time.Time) {
		fn(ctx, now)
	})
	job.With(t)

	// Queue job with the scheduler,
	// and store a new encompassing task.
	cncl := sch.sch.Schedule(job)
	sch.ts[id] = &task{
		job:  job,
		cncl: cncl,
	}

	return true
}

// task simply wraps together a scheduled
// job, and the matching cancel function.
type task struct {
	job  *sched.Job
	cncl func()
}
