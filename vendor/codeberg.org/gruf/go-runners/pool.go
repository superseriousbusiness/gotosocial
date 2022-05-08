package runners

import (
	"context"
	"runtime"
	"sync"
)

// WorkerFunc represents a function processable by a worker in WorkerPool. Note
// that implementations absolutely MUST check whether passed context is Done()
// otherwise stopping the pool may block for large periods of time.
type WorkerFunc func(context.Context)

// WorkerPool provides a means of enqueuing asynchronous work.
type WorkerPool struct {
	queue chan WorkerFunc
	free  chan struct{}
	wait  sync.WaitGroup
	svc   Service
}

// NewWorkerPool returns a new WorkerPool with provided worker count and WorkerFunc queue size.
// The number of workers represents how many WorkerFuncs can be executed simultaneously, and the
// queue size represents the max number of WorkerFuncs that can be queued at any one time.
func NewWorkerPool(workers int, queue int) WorkerPool {
	if workers < 1 {
		workers = runtime.GOMAXPROCS(0)
	}
	if queue < 1 {
		queue = workers * 2
	}
	return WorkerPool{
		queue: make(chan WorkerFunc, queue),
		free:  make(chan struct{}, workers),
	}
}

// Start will attempt to start the worker pool, asynchronously. Return is success state.
func (pool *WorkerPool) Start() bool {
	ok := true

	done := make(chan struct{})
	go func() {
		ok = pool.svc.Run(func(ctx context.Context) {
			close(done)
			pool.process(ctx)
		})
		if !ok {
			close(done)
		}
	}()
	<-done

	return ok
}

// Stop will attempt to stop the worker pool, this will block until stopped. Return is success state.
func (pool *WorkerPool) Stop() bool {
	return pool.svc.Stop()
}

// Running returns whether the worker pool is running.
func (pool *WorkerPool) Running() bool {
	return pool.svc.Running()
}

// execute will take a queued function and pass it to a free worker when available.
func (pool *WorkerPool) execute(ctx context.Context, fn WorkerFunc) {
	var acquired bool

	// Set as running
	pool.wait.Add(1)

	select {
	// Pool context cancelled
	// (we fall through and let
	// the function execute).
	case <-ctx.Done():

	// Free worker acquired.
	case pool.free <- struct{}{}:
		acquired = true
	}

	go func() {
		defer func() {
			// defer in case panic
			if acquired {
				<-pool.free
			}
			pool.wait.Done()
		}()

		// Run queued
		fn(ctx)
	}()
}

// process is the background processing routine that passes queued functions to workers.
func (pool *WorkerPool) process(ctx context.Context) {
	for {
		select {
		// Pool context cancelled
		case <-ctx.Done():
			for {
				select {
				// Pop and execute queued
				case fn := <-pool.queue:
					fn(ctx) // ctx is closed

				// Empty, wait for workers
				default:
					pool.wait.Wait()
					return
				}
			}

		// Queued func received
		case fn := <-pool.queue:
			pool.execute(ctx, fn)
		}
	}
}

// Enqueue will add provided WorkerFunc to the queue to be performed when there is a free worker.
// This will block until the function has been queued. 'fn' will ALWAYS be executed, even on pool
// close, which can be determined via context <-ctx.Done(). WorkerFuncs MUST respect the passed context.
func (pool *WorkerPool) Enqueue(fn WorkerFunc) {
	// Check valid fn
	if fn == nil {
		return
	}

	select {
	// Pool context cancelled
	case <-pool.svc.Done():
		fn(closedctx)

	// Placed fn in queue
	case pool.queue <- fn:
	}
}

// EnqueueNoBlock attempts Enqueue but returns false if not executed.
func (pool *WorkerPool) EnqueueNoBlock(fn WorkerFunc) bool {
	// Check valid fn
	if fn == nil {
		return false
	}

	select {
	// Pool context cancelled
	case <-pool.svc.Done():
		return false

	// Placed fn in queue
	case pool.queue <- fn:
		return true

	// Queue is full
	default:
		return false
	}
}

// Queue returns the number of currently queued WorkerFuncs.
func (pool *WorkerPool) Queue() int {
	return len(pool.queue)
}

// Workers returns the number of currently active workers.
func (pool *WorkerPool) Workers() int {
	return len(pool.free)
}
