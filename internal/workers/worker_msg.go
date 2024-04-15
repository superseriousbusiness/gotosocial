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

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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
func (p *MsgWorkerPool[T]) Start(n int) (ok bool) {
	if ok = (len(p.workers) == 0); ok {
		p.workers = make([]*MsgWorker[T], n)
		for i := range p.workers {
			p.workers[i] = new(MsgWorker[T])
			p.workers[i].Process = p.Process
			p.workers[i].Queue = &p.Queue
			ok = p.workers[i].Start() && ok
		}
	}
	return
}

// Stop will attempt to stop contained Worker{}s.
func (p *MsgWorkerPool[T]) Stop() (ok bool) {
	if ok = (len(p.workers) > 0); ok {
		for i := range p.workers {
			ok = p.workers[i].Stop() && ok
			p.workers[i] = nil
		}
		p.workers = p.workers[:0]
	}
	return
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
	log.Infof(ctx, "%p: starting worker", w)
	defer log.Infof(ctx, "%p: stopped worker", w)
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

loop:
	for {
		select {
		// Worker ctx done.
		case <-ctx.Done():
			return

		// New message enqueued!
		case <-w.Queue.Wait():
		}

		// Try pop next message.
		msg, ok := w.Queue.Pop()
		if !ok {
			continue loop
		}

		// Attempt to process popped message type.
		if err := w.Process(ctx, msg); err != nil {
			log.Errorf(ctx, "%p: error processing: %v", w, err)
		}
	}
}
