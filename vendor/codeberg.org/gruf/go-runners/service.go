package runners

import (
	"context"
	"sync"
)

// Service provides a means of tracking a single long-running service, provided protected state
// changes and preventing multiple instances running. Also providing service state information.
type Service struct {
	state uint32             // 0=stopped, 1=running, 2=stopping
	wait  sync.Mutex         // wait is the mutex used as a single-entity wait-group, i.e. just a "wait" :p
	cncl  context.CancelFunc // cncl is the cancel function set for the current context
	ctx   context.Context    // ctx is the current context for running function (or nil if not running)
	mu    sync.Mutex         // mu protects state changes
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
		svc.Stop()
	}()

	// Run user func
	if fn != nil {
		fn(ctx)
	}
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
			svc.Stop()
		}()

		// Run user func
		if fn != nil {
			fn(ctx)
		}
	}()

	return true
}

// Stop will attempt to stop the service, cancelling the running function's context. Immediately
// returns false if not running, and true only after Service is fully stopped.
func (svc *Service) Stop() bool {
	// Attempt to stop the svc
	cncl, ok := svc.doStop()
	if !ok {
		return false
	}

	defer func() {
		// Get svc lock
		svc.mu.Lock()

		// Wait until stopped
		svc.wait.Lock()
		svc.wait.Unlock()

		// Reset the svc
		svc.ctx = nil
		svc.cncl = nil
		svc.state = 0
		svc.mu.Unlock()
	}()

	cncl() // cancel ctx
	return true
}

// doStart will safely set Service state to started, returning a ptr to this context insance.
func (svc *Service) doStart() (context.Context, bool) {
	// Protect startup
	svc.mu.Lock()

	if svc.state != 0 /* not stopped */ {
		svc.mu.Unlock()
		return nil, false
	}

	// state started
	svc.state = 1

	// Take our own ptr
	var ctx context.Context

	if svc.ctx == nil {
		// Context required allocating
		svc.ctx, svc.cncl = ContextWithCancel()
	}

	// Start the waiter
	svc.wait.Lock()

	// Set our ptr + unlock
	ctx = svc.ctx
	svc.mu.Unlock()

	return ctx, true
}

// doStop will safely set Service state to stopping, returning a ptr to this cancelfunc instance.
func (svc *Service) doStop() (context.CancelFunc, bool) {
	// Protect stop
	svc.mu.Lock()

	if svc.state != 1 /* not started */ {
		svc.mu.Unlock()
		return nil, false
	}

	// state stopping
	svc.state = 2

	// Take our own ptr
	// and unlock state
	cncl := svc.cncl
	svc.mu.Unlock()

	return cncl, true
}

// Running returns if Service is running (i.e. state NOT stopped / stopping).
func (svc *Service) Running() bool {
	svc.mu.Lock()
	state := svc.state
	svc.mu.Unlock()
	return (state == 1)
}

// Done returns a channel that's closed when Service.Stop() is called. It is
// the same channel provided to the currently running service function.
func (svc *Service) Done() <-chan struct{} {
	var done <-chan struct{}

	svc.mu.Lock()
	switch svc.state {
	// stopped
	// (here we create a new context so that the
	// returned 'done' channel here will still
	// be valid for when Service is next started)
	case 0:
		if svc.ctx == nil {
			// need to allocate new context
			svc.ctx, svc.cncl = ContextWithCancel()
		}
		done = svc.ctx.Done()

	// started
	case 1:
		done = svc.ctx.Done()

	// stopping
	case 2:
		done = svc.ctx.Done()
	}
	svc.mu.Unlock()

	return done
}
