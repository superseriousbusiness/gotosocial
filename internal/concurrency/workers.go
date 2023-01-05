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

package concurrency

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-runners"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// WorkerPool represents a proccessor for MsgType objects, using a worker pool to allocate resources.
type WorkerPool[MsgType any] struct {
	workers runners.WorkerPool
	process func(context.Context, MsgType) error
	nw, nq  int
	wtype   string // contains worker type for logging
}

// New returns a new WorkerPool[MsgType] with given number of workers and queue ratio,
// where the queue ratio is multiplied by no. workers to get queue size. If args < 1
// then suitable defaults are determined from the runtime's GOMAXPROCS variable.
func NewWorkerPool[MsgType any](workers int, queueRatio int) *WorkerPool[MsgType] {
	var zero MsgType

	if workers < 1 {
		// ensure sensible workers
		workers = runtime.GOMAXPROCS(0) * 4
	}
	if queueRatio < 1 {
		// ensure sensible ratio
		queueRatio = 100
	}

	// Calculate the short type string for the msg type
	msgType := reflect.TypeOf(zero).String()
	_, msgType = path.Split(msgType)

	w := &WorkerPool[MsgType]{
		process: nil,
		nw:      workers,
		nq:      workers * queueRatio,
		wtype:   fmt.Sprintf("worker.Worker[%s]", msgType),
	}

	// Log new worker creation with worker type prefix
	log.Infof("%s created with workers=%d queue=%d",
		w.wtype,
		workers,
		workers*queueRatio,
	)

	return w
}

// Start will attempt to start the underlying worker pool, or return error.
func (w *WorkerPool[MsgType]) Start() error {
	log.Infof("%s starting", w.wtype)

	// Check processor was set
	if w.process == nil {
		return errors.New("nil Worker.process function")
	}

	// Attempt to start pool
	if !w.workers.Start(w.nw, w.nq) {
		return errors.New("failed to start Worker pool")
	}

	return nil
}

// Stop will attempt to stop the underlying worker pool, or return error.
func (w *WorkerPool[MsgType]) Stop() error {
	log.Infof("%s stopping", w.wtype)

	// Attempt to stop pool
	if !w.workers.Stop() {
		return errors.New("failed to stop Worker pool")
	}

	return nil
}

// SetProcessor will set the Worker's processor function, which is called for each queued message.
func (w *WorkerPool[MsgType]) SetProcessor(fn func(context.Context, MsgType) error) {
	if w.process != nil {
		log.Panicf("%s Worker.process is already set", w.wtype)
	}
	w.process = fn
}

// Queue will queue provided message to be processed with there's a free worker.
func (w *WorkerPool[MsgType]) Queue(msg MsgType) {
	log.Tracef("%s queueing message: %+v", w.wtype, msg)

	// Create new process function for msg
	process := func(ctx context.Context) {
		if err := w.process(ctx, msg); err != nil {
			log.WithFields(kv.Fields{
				kv.Field{K: "type", V: w.wtype},
				kv.Field{K: "error", V: err},
			}...).Error("message processing error")
		}
	}

	// Attempt a fast-enqueue of process
	if !w.workers.EnqueueNow(process) {
		// No spot acquired, log warning
		log.WithFields(kv.Fields{
			kv.Field{K: "type", V: w.wtype},
			kv.Field{K: "queue", V: w.workers.Queue()},
		}...).Warn("full worker queue")

		// Block on enqueuing process func
		w.workers.Enqueue(process)
	}
}
