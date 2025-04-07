//go:build go1.22 && !go1.25

package mutexes

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// syncCond_last_ticket is an unsafe function that returns
// the ticket of the last awoken / notified goroutine by a
// a sync.Cond{}. it relies on expected memory layout.
func syncCond_last_ticket(c *sync.Cond) uint32 {

	// NOTE: must remain in
	// sync with runtime.notifyList{}.
	//
	// goexperiment.staticlockranking
	// does change it slightly, but
	// this does not alter the first
	// 2 fields which are all we need.
	type notifyList struct {
		_      atomic.Uint32
		notify uint32
		// ... other fields
	}

	// NOTE: must remain in
	// sync with sync.Cond{}.
	type syncCond struct {
		_ struct{}
		L sync.Locker
		n notifyList
		// ... other fields
	}

	// This field must be atomcially accessed.
	cptr := (*syncCond)(unsafe.Pointer(c))
	return atomic.LoadUint32(&cptr.n.notify)
}
