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

package workers

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/queue"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-runners"
)

// FnWorkerPool wraps multiple FnWorker{}s in
// a singular struct for easy multi start / stop.
type FnWorkerPool struct {

	// Queue is embedded queue.SimpleQueue{}
	// passed to each of the pool Worker{}s.
	Queue queue.SimpleQueue[func(context.Context)]

	// internal fields.
	workers []*FnWorker
}

// Start will attempt to start 'n' FnWorker{}s.
func (p *FnWorkerPool) Start(n int) {
	// Check whether workers are
	// set (is already running).
	ok := (len(p.workers) > 0)
	if ok {
		return
	}

	// Allocate new workers slice.
	p.workers = make([]*FnWorker, n)
	for i := range p.workers {

		// Allocate new FnWorker{}.
		p.workers[i] = new(FnWorker)
		p.workers[i].Queue = &p.Queue

		// Attempt to start worker.
		// Return bool not useful
		// here, as true = started,
		// false = already running.
		_ = p.workers[i].Start()
	}
}

// Stop will attempt to stop contained FnWorker{}s.
func (p *FnWorkerPool) Stop() {
	// Check whether workers are
	// set (is currently running).
	ok := (len(p.workers) == 0)
	if ok {
		return
	}

	// Stop all running workers.
	for i := range p.workers {

		// return bool not useful
		// here, as true = stopped,
		// false = never running.
		_ = p.workers[i].Stop()
	}

	// Unset workers slice.
	p.workers = p.workers[:0]
}

// FnWorker wraps a queue.SimpleQueue{} which
// it feeds from to provide it with function
// tasks to execute. It does so in a single
// goroutine with state management utilities.
type FnWorker struct {

	// Queue is the fn queue that FnWorker
	// will feed from for upcoming tasks.
	Queue *queue.SimpleQueue[func(context.Context)]

	// internal fields.
	service runners.Service
}

// Start will attempt to start the Worker{}.
func (w *FnWorker) Start() bool {
	return w.service.GoRun(w.run)
}

// Stop will attempt to stop the Worker{}.
func (w *FnWorker) Stop() bool {
	return w.service.Stop()
}

// run wraps process to restart on any panic.
func (w *FnWorker) run(ctx context.Context) {
	if w.Queue == nil {
		panic("not yet initialized")
	}
	log.Debugf(ctx, "%p: starting worker", w)
	defer log.Debugf(ctx, "%p: stopped worker", w)
	util.Must(func() { w.process(ctx) })
}

// process is the main delivery worker processing routine.
func (w *FnWorker) process(ctx context.Context) {
	if w.Queue == nil {
		// we perform this check here just
		// to ensure the compiler knows these
		// variables aren't nil in the loop,
		// even if already checked by caller.
		panic("not yet initialized")
	}

	for {
		// Block until pop next func.
		fn, ok := w.Queue.PopCtx(ctx)
		if !ok {
			return
		}

		// run!
		fn(ctx)
	}
}
