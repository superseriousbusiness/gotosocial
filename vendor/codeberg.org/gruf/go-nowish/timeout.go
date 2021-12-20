package nowish

import (
	"sync"
	"sync/atomic"
	"time"
)

// Timeout provides a reusable structure for enforcing timeouts with a cancel.
type Timeout struct {
	timer *time.Timer // timer is the underlying timeout-timer
	cncl  syncer      // cncl is the cancel synchronization channel
	next  int64       // next is the next timeout duration to run on
	state uint32      // state stores the current timeout state
	mu    sync.Mutex  // mu protects state, and helps synchronize return of .Start()
}

// NewTimeout returns a new Timeout instance.
func NewTimeout() Timeout {
	timer := time.NewTimer(time.Minute)
	timer.Stop() // don't keep it running
	return Timeout{
		timer: timer,
		cncl:  make(syncer),
	}
}

// startTimeout is the main timeout routine, handling starting the
// timeout runner at first and upon any time extensions, and handling
// any received cancels by stopping the running timer.
func (t *Timeout) startTimeout(hook func()) {
	var cancelled bool

	// Receive first timeout duration
	d := atomic.SwapInt64(&t.next, 0)

	// Indicate finished starting, this
	// was left locked by t.start().
	t.mu.Unlock()

	for {
		// Run supplied timeout
		cancelled = t.runTimeout(d)
		if cancelled {
			break
		}

		// Check for extension or set timed out
		d = atomic.SwapInt64(&t.next, 0)
		if d < 1 {
			if t.timedOut() {
				// timeout reached
				hook()
				break
			} else {
				// already cancelled
				t.cncl.wait()
				cancelled = true
				break
			}
		}

		if !t.extend() {
			// already cancelled
			t.cncl.wait()
			cancelled = true
			break
		}
	}

	if cancelled {
		// Release the .Cancel()
		defer t.cncl.notify()
	}

	// Mark as done
	t.reset()
}

// runTimeout will until supplied timeout or cancel called.
func (t *Timeout) runTimeout(d int64) (cancelled bool) {
	// Start the timer for 'd'
	t.timer.Reset(time.Duration(d))

	select {
	// Timeout reached
	case <-t.timer.C:
		if !t.timingOut() {
			// a sneaky cancel!
			t.cncl.wait()
			cancelled = true
		}

	// Cancel called
	case <-t.cncl.wait():
		cancelled = true
		if !t.timer.Stop() {
			<-t.timer.C
		}
	}

	return cancelled
}

// Start starts the timer with supplied timeout. If timeout is reached before
// cancel then supplied timeout hook will be called. Panic will be called if
// Timeout is already running when calling this function.
func (t *Timeout) Start(d time.Duration, hook func()) {
	if !t.start() {
		t.mu.Unlock() // need to unlock
		panic("timeout already started")
	}

	// Start the timeout
	atomic.StoreInt64(&t.next, int64(d))
	go t.startTimeout(hook)

	// Wait until start
	t.mu.Lock()
	t.mu.Unlock()
}

// Extend will attempt to extend the timeout runner's time, returns false if not running.
func (t *Timeout) Extend(d time.Duration) bool {
	var ok bool
	if ok = t.running(); ok {
		atomic.AddInt64(&t.next, int64(d))
	}
	return ok
}

// Cancel cancels the currently running timer. If a cancel is achieved, then
// this function will return after the timeout goroutine is finished.
func (t *Timeout) Cancel() {
	if !t.cancel() {
		return
	}
	t.cncl.notify()
	<-t.cncl.wait()
}

// possible timeout states.
const (
	stopped   = 0
	started   = 1
	timingOut = 2
	cancelled = 3
	timedOut  = 4
)

// cas will perform a compare and swap where the compare is a provided function.
func (t *Timeout) cas(check func(uint32) bool, swap uint32) bool {
	var cas bool

	t.mu.Lock()
	if cas = check(t.state); cas {
		t.state = swap
	}
	t.mu.Unlock()

	return cas
}

// start attempts to mark the timeout state as 'started', note DOES NOT unlock Timeout.mu.
func (t *Timeout) start() bool {
	var ok bool

	t.mu.Lock()
	if ok = (t.state == stopped); ok {
		t.state = started
	}

	// don't unlock
	return ok
}

// timingOut attempts to mark the timeout state as 'timing out'.
func (t *Timeout) timingOut() bool {
	return t.cas(func(u uint32) bool {
		return (u == started)
	}, timingOut)
}

// timedOut attempts mark the 'timing out' state as 'timed out'.
func (t *Timeout) timedOut() bool {
	return t.cas(func(u uint32) bool {
		return (u == timingOut)
	}, timedOut)
}

// extend attempts to extend a 'timing out' state by moving it back to 'started'.
func (t *Timeout) extend() bool {
	return t.cas(func(u uint32) bool {
		return (u == started) ||
			(u == timingOut)
	}, started)
}

// running returns whether the state is anything other than 'stopped'.
func (t *Timeout) running() bool {
	t.mu.Lock()
	running := (t.state != stopped)
	t.mu.Unlock()
	return running
}

// cancel attempts to mark the timeout state as 'cancelled'.
func (t *Timeout) cancel() bool {
	return t.cas(func(u uint32) bool {
		return (u == started) ||
			(u == timingOut)
	}, cancelled)
}

// reset marks the timeout state as 'stopped'.
func (t *Timeout) reset() {
	t.mu.Lock()
	t.state = stopped
	t.mu.Unlock()
}

// syncer provides helpful receiver methods for a synchronization channel.
type syncer (chan struct{})

// notify blocks on sending an empty value down channel.
func (s syncer) notify() {
	s <- struct{}{}
}

// wait returns the underlying channel for blocking until '.notify()'.
func (s syncer) wait() <-chan struct{} {
	return s
}
