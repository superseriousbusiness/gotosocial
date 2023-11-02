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

// Scheduler wraps an underlying task scheduler
// to provide concurrency safe tracking by 'id'
// strings in order to provide easy cancellation.
type Scheduler struct {
	sch sched.Scheduler
	ts  map[string]*task
	mu  sync.Mutex
}

// Start will start the Scheduler background routine, returning success.
// Note that this creates a new internal task map, stopping and dropping
// all previously known running tasks.
func (sch *Scheduler) Start() bool {
	if sch.sch.Start(nil) {
		sch.ts = make(map[string]*task)
		return true
	}
	return false
}

// Stop will stop the Scheduler background routine, returning success.
// Note that this nils-out the internal task map, stopping and dropping
// all previously known running tasks.
func (sch *Scheduler) Stop() bool {
	if sch.sch.Stop() {
		sch.ts = nil
		return true
	}
	return false
}

// AddOnce adds a run-once job with given id, function and timing parameters, returning success.
func (sch *Scheduler) AddOnce(id string, start time.Time, fn func(context.Context, time.Time)) bool {
	return sch.schedule(id, fn, (*sched.Once)(&start))
}

// AddRecurring adds a new recurring job with given id, function and timing parameters, returning success.
func (sch *Scheduler) AddRecurring(id string, start time.Time, freq time.Duration, fn func(context.Context, time.Time)) bool {
	return sch.schedule(id, fn, &sched.PeriodicAt{Once: sched.Once(start), Period: sched.Periodic(freq)})
}

// Cancel will attempt to cancel job with given id,
// dropping it from internal scheduler and task map.
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

type task struct {
	job  *sched.Job
	cncl func()
}
