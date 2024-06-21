package mutexes

import (
	"sync"
)

// Cond is similar to a sync.Cond{}, but
// it encompasses the Mutex{} within itself.
type Cond struct {
	c sync.Cond
	sync.Mutex
}

// See: sync.Cond{}.Wait().
func (c *Cond) Wait() {
	if c.c.L == nil {
		c.c.L = &c.Mutex
	}
	c.c.Wait()
}

// See: sync.Cond{}.Signal().
func (c *Cond) Signal() {
	if c.c.L == nil {
		c.c.L = &c.Mutex
	}
	c.c.Signal()
}

// See: sync.Cond{}.Broadcast().
func (c *Cond) Broadcast() {
	if c.c.L == nil {
		c.c.L = &c.Mutex
	}
	c.c.Broadcast()
}

// RWCond is similar to a sync.Cond{}, but
// it encompasses the RWMutex{} within itself.
type RWCond struct {
	c sync.Cond
	sync.RWMutex
}

// See: sync.Cond{}.Wait().
func (c *RWCond) Wait() {
	if c.c.L == nil {
		c.c.L = &c.RWMutex
	}
	c.c.Wait()
}

// See: sync.Cond{}.Signal().
func (c *RWCond) Signal() {
	if c.c.L == nil {
		c.c.L = &c.RWMutex
	}
	c.c.Signal()
}

// See: sync.Cond{}.Broadcast().
func (c *RWCond) Broadcast() {
	if c.c.L == nil {
		c.c.L = &c.RWMutex
	}
	c.c.Broadcast()
}
