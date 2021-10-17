package nowish

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Start returns a new Clock instance initialized and
// started with the provided precision, along with the
// stop function for it's underlying timer
func Start(precision time.Duration) (*Clock, func()) {
	c := Clock{}
	return &c, c.Start(precision)
}

type Clock struct {
	noCopy noCopy //nolint noCopy because a copy will fuck with atomics

	// format stores the time formatting style string
	format string

	// valid indicates whether the current value stored in .Format is valid
	valid uint32

	// mutex protects writes to .Format, not because it would be unsafe, but
	// because we want to minimize unnnecessary allocations
	mutex sync.Mutex

	// nowfmt is an unsafe pointer to the last-updated time format string
	nowfmt unsafe.Pointer

	// now is an unsafe pointer to the last-updated time.Time object
	now unsafe.Pointer
}

// Start starts the clock with the provided precision, the
// returned function is the stop function for the underlying timer
func (c *Clock) Start(precision time.Duration) func() {
	// Create ticker from duration
	tick := time.NewTicker(precision)

	// Set initial time
	t := time.Now()
	atomic.StorePointer(&c.now, unsafe.Pointer(&t))

	// Set initial format
	s := ""
	atomic.StorePointer(&c.nowfmt, unsafe.Pointer(&s))

	// If formatting string unset, set default
	c.mutex.Lock()
	if c.format == "" {
		c.format = time.RFC822
	}
	c.mutex.Unlock()

	// Start main routine
	go c.run(tick)

	// Return stop fn
	return tick.Stop
}

// run is the internal clock ticking loop
func (c *Clock) run(tick *time.Ticker) {
	for {
		// Wait on tick
		_, ok := <-tick.C

		// Channel closed
		if !ok {
			break
		}

		// Update time
		t := time.Now()
		atomic.StorePointer(&c.now, unsafe.Pointer(&t))

		// Invalidate format string
		atomic.StoreUint32(&c.valid, 0)
	}
}

// Now returns a good (ish) estimate of the current 'now' time
func (c *Clock) Now() time.Time {
	return *(*time.Time)(atomic.LoadPointer(&c.now))
}

// NowFormat returns the formatted "now" time, cached until next tick and "now" updates
func (c *Clock) NowFormat() string {
	// If format still valid, return this
	if atomic.LoadUint32(&c.valid) == 1 {
		return *(*string)(atomic.LoadPointer(&c.nowfmt))
	}

	// Get mutex lock
	c.mutex.Lock()

	// Double check still invalid
	if atomic.LoadUint32(&c.valid) == 1 {
		c.mutex.Unlock()
		return *(*string)(atomic.LoadPointer(&c.nowfmt))
	}

	// Calculate time format
	b := c.Now().AppendFormat(
		make([]byte, 0, len(c.format)),
		c.format,
	)

	// Update the stored value and set valid!
	atomic.StorePointer(&c.nowfmt, unsafe.Pointer(&b))
	atomic.StoreUint32(&c.valid, 1)

	// Unlock and return
	c.mutex.Unlock()

	// Note:
	// it's safe to do this conversion here
	// because this byte slice will never change.
	// and we have the direct pointer to it, we're
	// not requesting it atomicly via c.Format
	return *(*string)(unsafe.Pointer(&b))
}

// SetFormat sets the time format string used by .NowFormat()
func (c *Clock) SetFormat(format string) {
	// Get mutex lock
	c.mutex.Lock()

	// Update time format
	c.format = format

	// Invalidate current format string
	atomic.StoreUint32(&c.valid, 0)

	// Unlock
	c.mutex.Unlock()
}
