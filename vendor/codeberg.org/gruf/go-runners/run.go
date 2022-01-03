package runners

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// FuncRunner provides a means of managing long-running functions e.g. main logic loops.
type FuncRunner struct {
	// HandOff is the time after which a blocking function will be considered handed off
	HandOff time.Duration

	// ErrorHandler is the function that errors are passed to when encountered by the
	// provided function. This can be used both for logging, and for error filtering
	ErrorHandler func(err error) error

	svc Service    // underlying service to manage start/stop
	err error      // last-set error
	mu  sync.Mutex // protects err
}

// Go will attempt to run 'fn' asynchronously. The provided context is used to propagate requested
// cancel if FuncRunner.Stop() is called. Any returned error will be passed to FuncRunner.ErrorHandler
// for filtering/logging/etc. Any blocking functions will be waited on for FuncRunner.HandOff amount of
// time before considering the function as handed off. Returned bool is success state, i.e. returns true
// if function is successfully handed off or returns within hand off time with nil error.
func (r *FuncRunner) Go(fn func(ctx context.Context) error) bool {
	done := make(chan struct{})

	go func() {
		var cancelled bool

		has := r.svc.Run(func(ctx context.Context) {
			// reset error
			r.mu.Lock()
			r.err = nil
			r.mu.Unlock()

			// Run supplied func and set errror if returned
			if err := Run(func() error { return fn(ctx) }); err != nil {
				r.mu.Lock()
				r.err = err
				r.mu.Unlock()
			}

			// signal done
			close(done)

			// Check if cancelled
			select {
			case <-ctx.Done():
				cancelled = true
			default:
				cancelled = false
			}
		})

		switch has {
		// returned after starting
		case true:
			r.mu.Lock()

			// filter out errors due FuncRunner.Stop() being called
			if cancelled && errors.Is(r.err, context.Canceled) {
				// filter out errors from FuncRunner.Stop() being called
				r.err = nil
			} else if r.err != nil && r.ErrorHandler != nil {
				// pass any non-nil error to set handler
				r.err = r.ErrorHandler(r.err)
			}

			r.mu.Unlock()

		// already running
		case false:
			close(done)
		}
	}()

	// get valid handoff to use
	handoff := r.HandOff
	if handoff < 1 {
		handoff = time.Second * 5
	}

	select {
	// handed off (long-run successful)
	case <-time.After(handoff):
		return true

	// 'fn' returned, check error
	case <-done:
		return (r.Err() == nil)
	}
}

// Stop will cancel the context supplied to the running function.
func (r *FuncRunner) Stop() bool {
	return r.svc.Stop()
}

// Err returns the last-set error value.
func (r *FuncRunner) Err() error {
	r.mu.Lock()
	err := r.err
	r.mu.Unlock()
	return err
}

// Run will execute the supplied 'fn' catching any panics. Returns either function-returned error or formatted panic.
func Run(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				// wrap and preserve existing error
				err = fmt.Errorf("caught panic: %w", e)
			} else {
				// simply create new error fromt iface
				err = fmt.Errorf("caught panic: %v", r)
			}
		}
	}()

	// run supplied func
	err = fn()
	return
}
