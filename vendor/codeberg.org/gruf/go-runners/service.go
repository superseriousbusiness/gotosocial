package runners

import (
	"context"
	"sync"
)

// Service provides a means of tracking a single long-running service, provided protected state
// changes and preventing multiple instances running. Also providing service state information.
type Service struct {
	state uint32     // 0=stopped, 1=running, 2=stopping
	mutex sync.Mutex // mutext protects overall state changes
	wait  sync.Mutex // wait is used as a single-entity wait-group, only ever locked within 'mutex'
	ctx   cancelctx  // ctx is the current context for running function (or nil if not running)
}

// Run will run the supplied function until completion, using given context to propagate cancel.
// Immediately returns false if the Service is already running, and true after completed run.
func (svc *Service) Run(fn func(context.Context)) bool {
	// Attempt to start the svc
	ctx, ok := svc.doStart()
	if !ok {
		return false
	}

	defer func() {
		// unlock single wait
		svc.wait.Unlock()

		// ensure stopped
		_ = svc.Stop()
	}()

	// Run
	fn(ctx)

	return true
}

// GoRun will run the supplied function until completion in a goroutine, using given context to
// propagate cancel. Immediately returns boolean indicating success, or that service is already running.
func (svc *Service) GoRun(fn func(context.Context)) bool {
	// Attempt to start the svc
	ctx, ok := svc.doStart()
	if !ok {
		return false
	}

	go func() {
		defer func() {
			// unlock single wait
			svc.wait.Unlock()

			// ensure stopped
			_ = svc.Stop()
		}()

		// Run
		fn(ctx)
	}()

	return true
}

// Stop will attempt to stop the service, cancelling the running function's context. Immediately
// returns false if not running, and true only after Service is fully stopped.
func (svc *Service) Stop() bool {
	// Attempt to stop the svc
	ctx, ok := svc.doStop()
	if !ok {
		return false
	}

	defer func() {
		// Get svc lock
		svc.mutex.Lock()

		// Wait until stopped
		svc.wait.Lock()
		svc.wait.Unlock()

		// Reset the svc
		svc.ctx = nil
		svc.state = 0
		svc.mutex.Unlock()
	}()

	// Cancel ctx
	close(ctx)

	return true
}

// While allows you to execute given function guaranteed within current
// service state. Please note that this will hold the underlying service
// state change mutex open while executing the function.
func (svc *Service) While(fn func()) {
	// Protect state change
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	// Run
	fn()
}

// doStart will safely set Service state to started, returning a ptr to this context insance.
func (svc *Service) doStart() (cancelctx, bool) {
	// Protect startup
	svc.mutex.Lock()

	if svc.state != 0 /* not stopped */ {
		svc.mutex.Unlock()
		return nil, false
	}

	// state started
	svc.state = 1

	if svc.ctx == nil {
		// this will only have been allocated
		// if svc.Done() was already called.
		svc.ctx = make(cancelctx)
	}

	// Start the waiter
	svc.wait.Lock()

	// Take our own ptr
	// and unlock state
	ctx := svc.ctx
	svc.mutex.Unlock()

	return ctx, true
}

// doStop will safely set Service state to stopping, returning a ptr to this cancelfunc instance.
func (svc *Service) doStop() (cancelctx, bool) {
	// Protect stop
	svc.mutex.Lock()

	if svc.state != 1 /* not started */ {
		svc.mutex.Unlock()
		return nil, false
	}

	// state stopping
	svc.state = 2

	// Take our own ptr
	// and unlock state
	ctx := svc.ctx
	svc.mutex.Unlock()

	return ctx, true
}

// Running returns if Service is running (i.e. state NOT stopped / stopping).
func (svc *Service) Running() bool {
	svc.mutex.Lock()
	state := svc.state
	svc.mutex.Unlock()
	return (state == 1)
}

// Done returns a channel that's closed when Service.Stop() is called. It is
// the same channel provided to the currently running service function.
func (svc *Service) Done() <-chan struct{} {
	var done <-chan struct{}

	svc.mutex.Lock()
	switch svc.state {
	// stopped
	case 0:
		if svc.ctx == nil {
			// here we create a new context so that the
			// returned 'done' channel here will still
			// be valid for when Service is next started.
			svc.ctx = make(cancelctx)
		}
		done = svc.ctx

	// started
	case 1:
		done = svc.ctx

	// stopping
	case 2:
		done = svc.ctx
	}
	svc.mutex.Unlock()

	return done
}
