package nowish

import (
	"sync/atomic"
	"time"
)

// Timeout provides a reusable structure for enforcing timeouts with a cancel
type Timeout struct {
	noCopy noCopy //nolint noCopy because a copy will mess with atomics

	tk *time.Ticker  // tk is the underlying timeout-timer
	ch chan struct{} // ch is the cancel propagation channel
	st timeoutState  // st stores the current timeout state (and protects concurrent use)
}

// NewTimeout returns a new Timeout instance
func NewTimeout() Timeout {
	tk := time.NewTicker(time.Minute)
	tk.Stop() // don't keep it running
	return Timeout{
		tk: tk,
		ch: make(chan struct{}),
	}
}

// Start starts the timer with supplied timeout. If timeout is reached before
// cancel then supplied timeout hook will be called. Error may be called if
// Timeout is already running when this function is called
func (t *Timeout) Start(d time.Duration, hook func()) {
	// Attempt to acquire start
	if !t.st.start() {
		panic("nowish: timeout already started")
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
}

// Cancel cancels the currently running timer. If a cancel is achieved, then
// this function will return after the timeout goroutine is finished
func (t *Timeout) Cancel() {
	// Attempt to acquire stop
	if !t.st.stop() {
		return
	}

	// Send a cancel signal
	t.ch <- struct{}{}
}

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
