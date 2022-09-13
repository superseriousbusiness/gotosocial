package runners

import (
	"context"
	"runtime"
	"sync"
)

// WorkerFunc represents a function processable by a worker in WorkerPool. Note
// that implementations absolutely MUST check whether passed context is <-ctx.Done()
// otherwise stopping the pool may block indefinitely.
type WorkerFunc func(context.Context)

// WorkerPool provides a means of enqueuing asynchronous work.
type WorkerPool struct {
	fns chan WorkerFunc
	svc Service
}

// Start will start the main WorkerPool management loop in a new goroutine, along
// with requested number of child worker goroutines. Returns false if already running.
func (pool *WorkerPool) Start(workers int, queue int) bool {
	// Attempt to start the svc
	ctx, ok := pool.svc.doStart()
	if !ok {
		return false
	}

	if workers < 1 {
		// Use $GOMAXPROCS as default worker count
		workers = runtime.GOMAXPROCS(0)
	}

	if queue < 0 {
		// Set a reasonable queue default
		queue = workers * 2
	}

	// Allocate pool queue of given size
	pool.fns = make(chan WorkerFunc, queue)

	go func() {
		defer func() {
			// unlock single wait
			pool.svc.wait.Unlock()

			// ensure stopped
			pool.svc.Stop()
		}()

		var wait sync.WaitGroup

		// Start goroutine worker functions
		for i := 0; i < workers; i++ {
			go func() {
				// Signal start/stop
				wait.Add(1)
				defer wait.Done()

				for {
					// Fetch next func from stack
					fn, ok := <-pool.fns
					if !ok {
						return
					}

					// Run with ctx
					fn(ctx)
				}
			}()
		}

		// Set GC finalizer to stop pool on dealloc
		runtime.SetFinalizer(pool, func(pool *WorkerPool) {
			pool.svc.Stop()
		})

		// Wait on ctx
		<-ctx.Done()

		// Stop all workers
		close(pool.fns)
		wait.Wait()
	}()

	return true
}

// Stop will stop the WorkerPool management loop, blocking until stopped.
func (pool *WorkerPool) Stop() bool {
	return pool.svc.Stop()
}

// Enqueue will add provided WorkerFunc to the queue to be performed when there is a free worker.
// This will block until function is queued or pool is stopped. In all cases, the WorkerFunc will be
// executed, with the state of the pool being indicated by <-ctx.Done() of the passed ctx.
// WorkerFuncs MUST respect the passed context.
func (pool *WorkerPool) Enqueue(fn WorkerFunc) {
	// Check valid fn
	if fn == nil {
		return
	}

	select {
	// Pool ctx cancelled
	case <-pool.svc.Done():
		fn(closedctx)

	// Placed fn in queue
	case pool.fns <- fn:
	}
}

// EnqueueCtx is functionally identical to WorkerPool.Enqueue() but returns early in the
// case that caller provided <-ctx.Done() is closed, WITHOUT running the WorkerFunc.
func (pool *WorkerPool) EnqueueCtx(ctx context.Context, fn WorkerFunc) {
	// Check valid fn
	if fn == nil {
		return
	}

	select {
	// Caller ctx cancelled
	case <-ctx.Done():

	// Pool ctx cancelled
	case <-pool.svc.Done():
		fn(closedctx)

	// Placed fn in queue
	case pool.fns <- fn:
	}
}

// EnqueueNow attempts Enqueue but returns false if not executed.
func (pool *WorkerPool) EnqueueNow(fn WorkerFunc) bool {
	// Check valid fn
	if fn == nil {
		return false
	}

	select {
	// Pool ctx cancelled
	case <-pool.svc.Done():
		return false

	// Placed fn in queue
	case pool.fns <- fn:
		return true

	// Queue is full
	default:
		return false
	}
}

// Queue returns the number of currently queued WorkerFuncs.
func (pool *WorkerPool) Queue() int {
	return len(pool.fns)
}
