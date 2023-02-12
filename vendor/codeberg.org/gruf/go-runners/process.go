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
	state uint32
	wait  sync.WaitGroup
	err   *error
}

// Process will process the given function if first-call, else blocking until
// the first function has returned, returning the same error result.
func (p *Processor) Process(proc Processable) (err error) {
	// Acquire state lock.
	p.mutex.Lock()

	if p.state != 0 {
		// Already running.
		//
		// Get current err ptr.
		errPtr := p.err

		// Wait until finish.
		p.mutex.Unlock()
		p.wait.Wait()
		return *errPtr
	}

	// Reset error ptr.
	p.err = new(error)

	// Set started.
	p.wait.Add(1)
	p.state = 1
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

		// Store error.
		*p.err = err

		// Mark done.
		p.wait.Done()

		// Set stopped.
		p.mutex.Lock()
		p.state = 0
		p.mutex.Unlock()
	}()

	// Run process.
	err = proc()
	return
}
