package worker

import (
	"context"
	"errors"
	"reflect"
	"runtime"

	"codeberg.org/gruf/go-runners"
	"github.com/sirupsen/logrus"
)

// Worker represents a proccessor for MsgType objects, using a worker pool to allocate resources.
type Worker[MsgType any] struct {
	workers runners.WorkerPool
	process func(context.Context, MsgType) error
	prefix  string
}

// New returns a new Worker[MsgType] with given number of workers and queue size
// (see runners.WorkerPool for more information on args). If args < 1 then suitable
// defaults are determined from the runtime's GOMAXPROCS variable.
func New[MsgType any](workers int, queue int) *Worker[MsgType] {
	if workers < 1 {
		// ensure sensible workers
		workers = runtime.GOMAXPROCS(0)
	}
	if queue < 1 {
		// ensure sensible queue
		queue = workers * 100
	}

	worker := Worker[MsgType]{
		workers: runners.NewWorkerPool(workers, queue),
		process: nil,
	}

	// Set worker type prefix and log creation
	worker.prefix = reflect.TypeOf(worker).String()
	logrus.Infof(worker.prefix+"created with workers=%d queue=%d", workers, queue)

	return &worker
}

// Start will attempt to start the underlying worker pool, or return error.
func (w *Worker[MsgType]) Start() error {
	logrus.Info(w.prefix + "starting")

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
func (w *Worker[MsgType]) Stop() error {
	logrus.Info(w.prefix + "stopping")

	// Attempt to stop pool
	if !w.workers.Stop() {
		return errors.New("failed to stop Worker pool")
	}

	return nil
}

// SetProcessor will set the Worker's processor function, which is called for each queued message.
func (w *Worker[MsgType]) SetProcessor(fn func(context.Context, MsgType) error) {
	if w.process != nil {
		logrus.Panic("Worker.process is already set")
	}
	w.process = fn
}

// Queue will queue provided message to be processed with there's a free worker.
func (w *Worker[MsgType]) Queue(msg MsgType) {
	logrus.Tracef(w.prefix+"queueing message: %+v", msg)
	w.workers.Enqueue(func(ctx context.Context) {
		if err := w.process(ctx, msg); err != nil {
			logrus.Error(err)
		}
	})
}
