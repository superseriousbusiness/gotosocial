package runners

import (
	"fmt"
	"sync"
)

// Processable defines a runnable process with error return
// that can be passed to a Processor instance for managed running.
type Processable func() error

// Processor acts similarly to a sync.Once object, except that it is reusable. After
// the first call to Process(), any further calls before this first has returned will
// block until the first call has returned, and return the same error. This ensures
// that only a single instance of it is ever running at any one time.
type Processor struct {
	mutex sync.Mutex
	wait  *sync.WaitGroup
	err   *error
}

// Process will process the given function if first-call, else blocking until
// the first function has returned, returning the same error result.
func (p *Processor) Process(proc Processable) (err error) {
	// Acquire state lock.
	p.mutex.Lock()

	if p.wait != nil {
		// Already running.
		//
		// Get current ptrs.
		waitPtr := p.wait
		errPtr := p.err

		// Free state lock.
		p.mutex.Unlock()

		// Wait for finish.
		waitPtr.Wait()
		return *errPtr
	}

	// Alloc waiter for new process.
	var wait sync.WaitGroup

	// No need to alloc new error as
	// we use the alloc'd named error
	// return required for panic handling.

	// Reset ptrs.
	p.wait = &wait
	p.err = &err

	// Set started.
	wait.Add(1)
	p.mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			if err != nil {
				rOld := r // wrap the panic so we don't lose existing returned error
				r = fmt.Errorf("panic occured after error %q: %v", err.Error(), rOld)
			}

			// Catch any panics and wrap as error.
			err = fmt.Errorf("caught panic: %v", r)
		}

		// Mark done.
		wait.Done()

		// Set stopped.
		p.mutex.Lock()
		p.wait = nil
		p.mutex.Unlock()
	}()

	// Run process.
	err = proc()
	return
}
