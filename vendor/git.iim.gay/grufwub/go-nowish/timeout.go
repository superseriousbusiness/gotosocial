package nowish

import (
	"errors"
	"sync/atomic"
	"time"
)

// ErrTimeoutStarted is returned if a Timeout interface is attempted to be reused while still in operation
var ErrTimeoutStarted = errors.New("nowish: timeout already started")

// timeoutState provides a thread-safe timeout state mechanism
type timeoutState uint32

// start attempts to start the state, must be already reset, returns success
func (t *timeoutState) start() bool {
	return atomic.CompareAndSwapUint32((*uint32)(t), 0, 1)
}

// stop attempts to stop the state, must already be started, returns success
func (t *timeoutState) stop() bool {
	return atomic.CompareAndSwapUint32((*uint32)(t), 1, 2)
}

// reset is fairly self explanatory
func (t *timeoutState) reset() {
	atomic.StoreUint32((*uint32)(t), 0)
}

// Timeout provides a reusable structure for enforcing timeouts with a cancel
type Timeout interface {
	// Start starts the timer with supplied timeout. If timeout is reached before
	// cancel then supplied timeout hook will be called. Error may be called if
	// Timeout is already running when this function is called
	Start(time.Duration, func()) error

	// Cancel cancels the currently running timer. If a cancel is achieved, then
	// this function will return after the timeout goroutine is finished
	Cancel()
}

// NewTimeout returns a new Timeout instance
func NewTimeout() Timeout {
	t := &timeout{
		tk: time.NewTicker(time.Minute),
		ch: make(chan struct{}),
	}
	t.tk.Stop() // don't keep it running
	return t
}

// timeout is the Timeout implementation that we force
// initialization on via NewTimeout by unexporting it
type timeout struct {
	noCopy noCopy //nolint noCopy because a copy will mess with atomics

	tk *time.Ticker  // tk is the underlying timeout-timer
	ch chan struct{} // ch is the cancel propagation channel
	st timeoutState  // st stores the current timeout state (and protects concurrent use)
}

func (t *timeout) Start(d time.Duration, hook func()) error {
	// Attempt to acquire start
	if !t.st.start() {
		return ErrTimeoutStarted
	}

	// Start the ticker
	t.tk.Reset(d)

	go func() {
		cancelled := false

		select {
		// Timeout reached
		case <-t.tk.C:
			if !t.st.stop() {
				// cancel was called in the nick of time
				<-t.ch
				cancelled = true
			}

		// Cancel called
		case <-t.ch:
			cancelled = true
		}

		// Stop ticker
		t.tk.Stop()

		// If timed out call hook
		if !cancelled {
			hook()
		}

		// Finally, reset state
		t.st.reset()
	}()

	return nil
}

func (t *timeout) Cancel() {
	// Attempt to acquire stop
	if !t.st.stop() {
		return
	}

	// Send a cancel signal
	t.ch <- struct{}{}
}
