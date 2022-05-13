package concurrency

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"

	"codeberg.org/gruf/go-runners"
	"github.com/sirupsen/logrus"
)

// WorkerPool represents a proccessor for MsgType objects, using a worker pool to allocate resources.
type WorkerPool[MsgType any] struct {
	workers runners.WorkerPool
	process func(context.Context, MsgType) error
	prefix  string // contains type prefix for logging
}

<<<<<<< HEAD:internal/concurrency/workers.go
// New returns a new Worker[MsgType] with given number of workers and queue size
// (see runners.WorkerPool for more information on args). If args < 1 then suitable
// defaults are determined from the runtime's GOMAXPROCS variable.
func NewWorkerPool[MsgType any](workers int, queue int) *WorkerPool[MsgType] {
=======
// New returns a new Worker[MsgType] with given number of workers and queue ratio,
// where the queue ratio is multiplied by no. workers to get queue size. If args < 1
// then suitable defaults are determined from the runtime's GOMAXPROCS variable.
func New[MsgType any](workers int, queueRatio int) *Worker[MsgType] {
	var zero MsgType

>>>>>>> upstream/main:internal/worker/workers.go
	if workers < 1 {
		// ensure sensible workers
		workers = runtime.GOMAXPROCS(0)
	}
	if queueRatio < 1 {
		// ensure sensible ratio
		queueRatio = 100
	}

<<<<<<< HEAD:internal/concurrency/workers.go
	w := &WorkerPool[MsgType]{
		workers: runners.NewWorkerPool(workers, queue),
		process: nil,
		prefix:  reflect.TypeOf(WorkerPool[MsgType]{}).String(), //nolint
=======
	// Calculate the short type string for the msg type
	msgType := reflect.TypeOf(zero).String()
	_, msgType = path.Split(msgType)

	w := &Worker[MsgType]{
		workers: runners.NewWorkerPool(workers, workers*queueRatio),
		process: nil,
		prefix:  fmt.Sprintf("worker.Worker[%s]", msgType),
>>>>>>> upstream/main:internal/worker/workers.go
	}

	// Log new worker creation with type prefix
	logrus.Infof("%s created with workers=%d queue=%d",
		w.prefix,
		workers,
		workers*queueRatio,
	)

	return w
}

// Start will attempt to start the underlying worker pool, or return error.
<<<<<<< HEAD:internal/concurrency/workers.go
func (w *WorkerPool[MsgType]) Start() error {
	logrus.Info(w.prefix, "starting")
=======
func (w *Worker[MsgType]) Start() error {
	logrus.Infof("%s starting", w.prefix)
>>>>>>> upstream/main:internal/worker/workers.go

	// Check processor was set
	if w.process == nil {
		return errors.New("nil Worker.process function")
	}

	// Attempt to start pool
	if !w.workers.Start() {
		return errors.New("failed to start Worker pool")
	}

	return nil
}

// Stop will attempt to stop the underlying worker pool, or return error.
<<<<<<< HEAD:internal/concurrency/workers.go
func (w *WorkerPool[MsgType]) Stop() error {
	logrus.Info(w.prefix, "stopping")
=======
func (w *Worker[MsgType]) Stop() error {
	logrus.Infof("%s stopping", w.prefix)
>>>>>>> upstream/main:internal/worker/workers.go

	// Attempt to stop pool
	if !w.workers.Stop() {
		return errors.New("failed to stop Worker pool")
	}

	return nil
}

// SetProcessor will set the Worker's processor function, which is called for each queued message.
func (w *WorkerPool[MsgType]) SetProcessor(fn func(context.Context, MsgType) error) {
	if w.process != nil {
		logrus.Panicf("%s Worker.process is already set", w.prefix)
	}
	w.process = fn
}

// Queue will queue provided message to be processed with there's a free worker.
<<<<<<< HEAD:internal/concurrency/workers.go
func (w *WorkerPool[MsgType]) Queue(msg MsgType) {
	logrus.Tracef("%s queueing message: %+v", w.prefix, msg)
=======
func (w *Worker[MsgType]) Queue(msg MsgType) {
	logrus.Tracef("%s queueing message (workers=%d queue=%d): %+v",
		w.prefix, w.workers.Workers(), w.workers.Queue(), msg,
	)
>>>>>>> upstream/main:internal/worker/workers.go
	w.workers.Enqueue(func(ctx context.Context) {
		if err := w.process(ctx, msg); err != nil {
			logrus.Errorf("%s %v", w.prefix, err)
		}
	})
}
