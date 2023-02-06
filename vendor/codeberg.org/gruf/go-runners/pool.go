package runners

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"

	"codeberg.org/gruf/go-errors/v2"
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

	if workers <= 0 {
		// Use $GOMAXPROCS as default.
		workers = runtime.GOMAXPROCS(0)
	}

	if queue < 0 {
		// Use reasonable queue default.
		queue = workers * 10
	}

	// Allocate pool queue of given size.
	//
	// This MUST be set BEFORE we return and NOT in
	// the launched goroutine, or there is a risk that
	// the pool may appear as closed for a short time
	// until the main goroutine has been entered.
	fns := make(chan WorkerFunc, queue)
	pool.fns = fns

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
			wait.Add(1)

			go func() {
				defer wait.Done()

				// Run worker function (retry on panic)
				for !worker_run(CancelCtx(ctx), fns) {
				}
			}()
		}

		// Wait on ctx
		<-ctx

		// Drain function queue.
		//
		// All functions in the queue MUST be
		// run, so we pass them a closed context.
		//
		// This mainly allows us to block until
		// the function queue is empty, as worker
		// functions will also continue draining in
		// the background with the (now) closed ctx.
		for !drain_queue(fns) {
			// retry on panic
		}

		// Now the queue is empty, we can
		// safely close the channel signalling
		// all of the workers to return.
		close(fns)
		wait.Wait()
	}()

	return true
}

// Stop will stop the WorkerPool management loop, blocking until stopped.
func (pool *WorkerPool) Stop() bool {
	return pool.svc.Stop()
}

// Running returns if WorkerPool management loop is running (i.e. NOT stopped / stopping).
func (pool *WorkerPool) Running() bool {
	return pool.svc.Running()
}

// Done returns a channel that's closed when WorkerPool.Stop() is called. It is the same channel provided to the currently running worker functions.
func (pool *WorkerPool) Done() <-chan struct{} {
	return pool.svc.Done()
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
func (pool *WorkerPool) EnqueueCtx(ctx context.Context, fn WorkerFunc) bool {
	// Check valid fn
	if fn == nil {
		return false
	}

	select {
	// Caller ctx cancelled
	case <-ctx.Done():
		return false

	// Pool ctx cancelled
	case <-pool.svc.Done():
		return false

	// Placed fn in queue
	case pool.fns <- fn:
		return true
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
	var l int
	pool.svc.While(func() {
		l = len(pool.fns)
	})
	return l
}

// worker_run is the main worker routine, accepting functions from 'fns' until it is closed.
func worker_run(ctx context.Context, fns <-chan WorkerFunc) bool {
	defer func() {
		// Recover and drop any panic
		if r := recover(); r != nil {
			const msg = "worker_run: recovered panic: %v\n\n%s\n"
			fmt.Fprintf(os.Stderr, msg, r, errors.GetCallers(2, 10))
		}
	}()

	for {
		// Wait on next func
		fn, ok := <-fns
		if !ok {
			return true
		}

		// Run with ctx
		fn(ctx)
	}
}

// drain_queue will drain and run all functions in worker queue, passing in a closed context.
func drain_queue(fns <-chan WorkerFunc) bool {
	defer func() {
		// Recover and drop any panic
		if r := recover(); r != nil {
			const msg = "drain_queue: recovered panic: %v\n\n%s\n"
			fmt.Fprintf(os.Stderr, msg, r, errors.GetCallers(2, 10))
		}
	}()

	for {
		select {
		// Run with closed ctx
		case fn := <-fns:
			fn(closedctx)

		// Queue is empty
		default:
			return true
		}
	}
}
