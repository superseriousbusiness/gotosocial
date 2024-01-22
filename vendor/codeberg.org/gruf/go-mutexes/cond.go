package mutexes

import (
	"sync"
	"unsafe"
)

// Cond is similar to a sync.Cond{}, but
// it encompasses the Mutex{} within itself.
type Cond struct {
	notify notifyList
	sync.Mutex
}

// See: sync.Cond{}.Wait().
func (c *Cond) Wait() {
	t := runtime_notifyListAdd(&c.notify)
	c.Mutex.Unlock()
	runtime_notifyListWait(&c.notify, t)
	c.Mutex.Lock()
}

// See: sync.Cond{}.Signal().
func (c *Cond) Signal() { runtime_notifyListNotifyOne(&c.notify) }

// See: sync.Cond{}.Broadcast().
func (c *Cond) Broadcast() { runtime_notifyListNotifyAll(&c.notify) }

// RWCond is similar to a sync.Cond{}, but
// it encompasses the RWMutex{} within itself.
type RWCond struct {
	notify notifyList
	sync.RWMutex
}

// See: sync.Cond{}.Wait().
func (c *RWCond) Wait() {
	t := runtime_notifyListAdd(&c.notify)
	c.RWMutex.Unlock()
	runtime_notifyListWait(&c.notify, t)
	c.RWMutex.Lock()
}

// See: sync.Cond{}.Signal().
func (c *RWCond) Signal() { runtime_notifyListNotifyOne(&c.notify) }

// See: sync.Cond{}.Broadcast().
func (c *RWCond) Broadcast() { runtime_notifyListNotifyAll(&c.notify) }

// unused fields left
// un-named for safety.
type notifyList struct {
	_      uint32         // wait   uint32
	notify uint32         // notify uint32
	_      uintptr        // lock   mutex
	_      unsafe.Pointer // head   *sudog
	_      unsafe.Pointer // tail   *sudog
}

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListAdd sync.runtime_notifyListAdd
func runtime_notifyListAdd(l *notifyList) uint32

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListWait sync.runtime_notifyListWait
func runtime_notifyListWait(l *notifyList, t uint32)

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListNotifyOne sync.runtime_notifyListNotifyOne
func runtime_notifyListNotifyOne(l *notifyList)

// See runtime/sema.go for documentation.
//
//go:linkname runtime_notifyListNotifyAll sync.runtime_notifyListNotifyAll
func runtime_notifyListNotifyAll(l *notifyList)

// Ensure that sync and runtime agree on size of notifyList.
//
//go:linkname runtime_notifyListCheck sync.runtime_notifyListCheck
func runtime_notifyListCheck(size uintptr)
func init() {
	var n notifyList
	runtime_notifyListCheck(unsafe.Sizeof(n))
}
