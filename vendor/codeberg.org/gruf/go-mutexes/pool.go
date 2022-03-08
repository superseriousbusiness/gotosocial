package mutexes

// pool is a very simply memory pool.
type pool struct {
	current []interface{}
	victim  []interface{}
	alloc   func() interface{}
}

// Acquire will returns a sync.RWMutex from pool (or alloc new).
func (p *pool) Acquire() interface{} {
	// First try the current queue
	if l := len(p.current) - 1; l >= 0 {
		v := p.current[l]
		p.current = p.current[:l]
		return v
	}

	// Next try the victim queue.
	if l := len(p.victim) - 1; l >= 0 {
		v := p.victim[l]
		p.victim = p.victim[:l]
		return v
	}

	// Lastly, alloc new.
	return p.alloc()
}

// Release places a sync.RWMutex back in the pool.
func (p *pool) Release(v interface{}) {
	p.current = append(p.current, v)
}

// GC will clear out unused entries from the pool.
func (p *pool) GC() {
	current := p.current
	p.current = nil
	p.victim = current
}
