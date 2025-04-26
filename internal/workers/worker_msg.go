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
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/queue"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-structr"
)

// MsgWorkerPool wraps multiple MsgWorker{}s in
// a singular struct for easy multi start / stop.
type MsgWorkerPool[Msg any] struct {

	// Process handles queued message types.
	Process func(context.Context, Msg) error

	// Queue is embedded queue.StructQueue{}
	// passed to each of the pool Worker{}s.
	Queue queue.StructQueue[Msg]

	// internal fields.
	workers []*MsgWorker[Msg]
}

// Init will initialize the worker pool queue with given struct indices.
func (p *MsgWorkerPool[T]) Init(indices []structr.IndexConfig) {
	p.Queue.Init(structr.QueueConfig[T]{Indices: indices})
}

// Start will attempt to start 'n' Worker{}s.
func (p *MsgWorkerPool[T]) Start(n int) {
	// Check whether workers are
	// set (is already running).
	ok := (len(p.workers) > 0)
	if ok {
		return
	}

	// Allocate new msg workers slice.
	p.workers = make([]*MsgWorker[T], n)
	for i := range p.workers {

		// Allocate new MsgWorker[T]{}.
		p.workers[i] = new(MsgWorker[T])
		p.workers[i].Process = p.Process
		p.workers[i].Queue = &p.Queue

		// Attempt to start worker.
		// Return bool not useful
		// here, as true = started,
		// false = already running.
		_ = p.workers[i].Start()
	}
}

// Stop will attempt to stop contained Worker{}s.
func (p *MsgWorkerPool[T]) Stop() {
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

// MsgWorker wraps a processing function to
// feed from a queue.StructQueue{} for messages
// to process. It does so in a single goroutine
// with state management utilities.
type MsgWorker[Msg any] struct {

	// Process handles queued message types.
	Process func(context.Context, Msg) error

	// Queue is the Delivery{} message queue
	// that delivery worker will feed from.
	Queue *queue.StructQueue[Msg]

	// internal fields.
	service runners.Service
}

// Start will attempt to start the Worker{}.
func (w *MsgWorker[T]) Start() bool {
	return w.service.GoRun(w.run)
}

// Stop will attempt to stop the Worker{}.
func (w *MsgWorker[T]) Stop() bool {
	return w.service.Stop()
}

// run wraps process to restart on any panic.
func (w *MsgWorker[T]) run(ctx context.Context) {
	if w.Process == nil || w.Queue == nil {
		panic("not yet initialized")
	}
	util.Must(func() { w.process(ctx) })
}

// process is the main delivery worker processing routine.
func (w *MsgWorker[T]) process(ctx context.Context) {
	if w.Process == nil || w.Queue == nil {
		// we perform this check here just
		// to ensure the compiler knows these
		// variables aren't nil in the loop,
		// even if already checked by caller.
		panic("not yet initialized")
	}

	for {
		// Block until pop next message.
		msg, ok := w.Queue.PopCtx(ctx)
		if !ok {
			return
		}

		// Attempt to process message.
		err := w.Process(ctx, msg)
		if err != nil {
			log.Errorf(ctx, "%p: error processing: %v", w, err)

			if errors.Is(err, context.Canceled) &&
				ctx.Err() != nil {
				// In the case of our own context
				// being cancelled, push message
				// back onto queue for persisting.
				//
				// Note we specifically check against
				// context.Canceled here as it will
				// be faster than the mutex lock of
				// ctx.Err(), so gives an initial
				// faster check in the if-clause.
				w.Queue.Push(msg)
				break
			}
		}
	}
}
