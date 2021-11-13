package nowish

import (
	"sync"
	"sync/atomic"
	"time"
)

// Timeout provides a reusable structure for enforcing timeouts with a cancel
type Timeout struct {
	noCopy noCopy //nolint noCopy because a copy will mess with atomics

	tk *time.Timer    // tk is the underlying timeout-timer
	ch syncer         // ch is the cancel synchronization channel
	wg sync.WaitGroup // wg is the waitgroup to hold .Start() until timeout goroutine started
	st timeoutState   // st stores the current timeout state (and protects concurrent use)
}

// NewTimeout returns a new Timeout instance
func NewTimeout() Timeout {
	tk := time.NewTimer(time.Minute)
	tk.Stop() // don't keep it running
	return Timeout{
		tk: tk,
		ch: make(syncer),
	}
}

func (t *Timeout) runTimeout(hook func()) {
	t.wg.Add(1)
	go func() {
		cancelled := false

		// Signal started
		t.wg.Done()

		select {
		// Timeout reached
		case <-t.tk.C:
			if !t.st.stop() /* a sneaky cancel! */ {
				t.ch.recv()
				cancelled = true
				defer t.ch.send()
			}

		// Cancel called
		case <-t.ch:
			cancelled = true
			defer t.ch.send()
		}

		// Ensure timer stopped
		if cancelled && !t.tk.Stop() {
			<-t.tk.C
		}

		// Defer reset state
		defer t.st.reset()

		// If timed out call hook
		if !cancelled {
			hook()
		}
	}()
	t.wg.Wait()
}

// Start starts the timer with supplied timeout. If timeout is reached before
// cancel then supplied timeout hook will be called. Error may be called if
// Timeout is already running when this function is called
func (t *Timeout) Start(d time.Duration, hook func()) {
	if !t.st.start() {
		panic("nowish: timeout already started")
	}
	t.runTimeout(hook)
	t.tk.Reset(d)
}

// Cancel cancels the currently running timer. If a cancel is achieved, then
// this function will return after the timeout goroutine is finished
func (t *Timeout) Cancel() {
	if !t.st.stop() {
		return
	}
	t.ch.send()
	t.ch.recv()
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

// syncer provides helpful receiver methods for a synchronization channel
type syncer (chan struct{})

// send blocks on sending an empty value down channel
func (s syncer) send() {
	s <- struct{}{}
}

// recv blocks on receiving (and dropping) empty value from channel
func (s syncer) recv() {
	<-s
}
