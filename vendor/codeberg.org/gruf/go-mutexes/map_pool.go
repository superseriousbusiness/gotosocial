package mutexes

// pool is a very simply memory pool.
type pool struct {
	current []*rwmutex
	victim  []*rwmutex
}

// Acquire will returns a rwmutex from pool (or alloc new).
func (p *pool) Acquire() *rwmutex {
	// First try the current queue
	if l := len(p.current) - 1; l >= 0 {
		mu := p.current[l]
		p.current = p.current[:l]
		return mu
	}

	// Next try the victim queue.
	if l := len(p.victim) - 1; l >= 0 {
		mu := p.victim[l]
		p.victim = p.victim[:l]
		return mu
	}

	// Lastly, alloc new.
	return &rwmutex{}
}

// Release places a sync.RWMutex back in the pool.
func (p *pool) Release(mu *rwmutex) {
	p.current = append(p.current, mu)
}

// GC will clear out unused entries from the pool.
func (p *pool) GC() {
	current := p.current
	p.current = nil
	p.victim = current
}
