package mutexes

// rwmutexPool is a very simply memory rwmutexPool.
type rwmutexPool struct {
	current []*rwmutexState
	victim  []*rwmutexState
}

// Acquire will returns a rwmutexState from rwmutexPool (or alloc new).
func (p *rwmutexPool) Acquire() *rwmutexState {
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
	return &rwmutexState{}
}

// Release places a sync.rwmutexState back in the rwmutexPool.
func (p *rwmutexPool) Release(mu *rwmutexState) {
	p.current = append(p.current, mu)
}

// GC will clear out unused entries from the rwmutexPool.
func (p *rwmutexPool) GC() {
	current := p.current
	p.current = nil
	p.victim = current
}
