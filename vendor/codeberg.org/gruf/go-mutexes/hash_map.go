package mutexes

type hashmap struct {
	m map[string]*rwmutex
	n int
}

func (m *hashmap) init(cap int) {
	m.m = make(map[string]*rwmutex, cap)
	m.n = cap
}

func (m *hashmap) Get(key string) *rwmutex { return m.m[key] }

func (m *hashmap) Put(key string, mu *rwmutex) {
	m.m[key] = mu
	if n := len(m.m); n > m.n {
		m.n = n
	}
}

func (m *hashmap) Delete(key string) {
	delete(m.m, key)
}

func (m *hashmap) Compact() {
	// Noop when hashmap size
	// is too small to matter.
	if m.n < 2048 {
		return
	}

	// Difference between maximum map
	// size and the current map size.
	diff := m.n - len(m.m)

	// Maximum load factor before
	// runtime allocates new hmap:
	// maxLoad = 13 / 16
	//
	// So we apply the inverse/2, once
	// $maxLoad/2 % of hmap is empty we
	// compact the map to drop buckets.
	if 2*16*diff > m.n*13 {

		// Create new map only as big as required.
		m2 := make(map[string]*rwmutex, len(m.m))
		for k, v := range m.m {
			m2[k] = v
		}

		// Set new.
		m.m = m2
		m.n = len(m2)
	}
}
