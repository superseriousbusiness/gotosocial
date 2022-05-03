package cache

// LookupCfg is the LookupCache configuration.
type LookupCfg[OGKey, AltKey comparable, Value any] struct {
	// RegisterLookups is called on init to register lookups
	// within LookupCache's internal LookupMap
	RegisterLookups func(*LookupMap[OGKey, AltKey])

	// AddLookups is called on each addition to the cache, to
	// set any required additional key lookups for supplied item
	AddLookups func(*LookupMap[OGKey, AltKey], Value)

	// DeleteLookups is called on each eviction/invalidation of
	// an item in the cache, to remove any unused key lookups
	DeleteLookups func(*LookupMap[OGKey, AltKey], Value)
}

// LookupCache is a cache built on-top of TTLCache, providing multi-key
// lookups for items in the cache by means of additional lookup maps. These
// maps simply store additional keys => original key, with hook-ins to automatically
// call user supplied functions on adding an item, or on updating/deleting an
// item to keep the LookupMap up-to-date.
type LookupCache[OGKey, AltKey comparable, Value any] interface {
	Cache[OGKey, Value]

	// GetBy fetches a cached value by supplied lookup identifier and key
	GetBy(lookup string, key AltKey) (value Value, ok bool)

	// CASBy will attempt to perform a CAS operation on supplied lookup identifier and key
	CASBy(lookup string, key AltKey, cmp, swp Value) bool

	// SwapBy will attempt to perform a swap operation on supplied lookup identifier and key
	SwapBy(lookup string, key AltKey, swp Value) Value

	// HasBy checks if a value is cached under supplied lookup identifier and key
	HasBy(lookup string, key AltKey) bool

	// InvalidateBy invalidates a value by supplied lookup identifier and key
	InvalidateBy(lookup string, key AltKey) bool
}

type lookupTTLCache[OK, AK comparable, V any] struct {
	config LookupCfg[OK, AK, V]
	lookup LookupMap[OK, AK]
	TTLCache[OK, V]
}

// NewLookup returns a new initialized LookupCache.
func NewLookup[OK, AK comparable, V any](cfg LookupCfg[OK, AK, V]) LookupCache[OK, AK, V] {
	switch {
	case cfg.RegisterLookups == nil:
		panic("cache: nil lookups register function")
	case cfg.AddLookups == nil:
		panic("cache: nil lookups add function")
	case cfg.DeleteLookups == nil:
		panic("cache: nil delete lookups function")
	}
	c := lookupTTLCache[OK, AK, V]{config: cfg}
	c.TTLCache.Init()
	c.lookup.lookup = make(map[string]map[AK]OK)
	c.config.RegisterLookups(&c.lookup)
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.lookup.initd = true
	return &c
}

func (c *lookupTTLCache[OK, AK, V]) SetEvictionCallback(hook Hook[OK, V]) {
	if hook == nil {
		hook = emptyHook[OK, V]
	}
	c.TTLCache.SetEvictionCallback(func(key OK, value V) {
		hook(key, value)
		c.config.DeleteLookups(&c.lookup, value)
	})
}

func (c *lookupTTLCache[OK, AK, V]) SetInvalidateCallback(hook Hook[OK, V]) {
	if hook == nil {
		hook = emptyHook[OK, V]
	}
	c.TTLCache.SetInvalidateCallback(func(key OK, value V) {
		hook(key, value)
		c.config.DeleteLookups(&c.lookup, value)
	})
}

func (c *lookupTTLCache[OK, AK, V]) GetBy(lookup string, key AK) (V, bool) {
	c.Lock()
	origKey, ok := c.lookup.Get(lookup, key)
	if !ok {
		c.Unlock()
		var value V
		return value, false
	}
	v, ok := c.GetUnsafe(origKey)
	c.Unlock()
	return v, ok
}

func (c *lookupTTLCache[OK, AK, V]) Put(key OK, value V) bool {
	c.Lock()
	put := c.PutUnsafe(key, value)
	if put {
		c.config.AddLookups(&c.lookup, value)
	}
	c.Unlock()
	return put
}

func (c *lookupTTLCache[OK, AK, V]) Set(key OK, value V) {
	c.Lock()
	defer c.Unlock()
	c.SetUnsafe(key, value)
	c.config.AddLookups(&c.lookup, value)
}

func (c *lookupTTLCache[OK, AK, V]) CASBy(lookup string, key AK, cmp, swp V) bool {
	c.Lock()
	defer c.Unlock()
	origKey, ok := c.lookup.Get(lookup, key)
	if !ok {
		return false
	}
	return c.CASUnsafe(origKey, cmp, swp)
}

func (c *lookupTTLCache[OK, AK, V]) SwapBy(lookup string, key AK, swp V) V {
	c.Lock()
	defer c.Unlock()
	origKey, ok := c.lookup.Get(lookup, key)
	if !ok {
		var value V
		return value
	}
	return c.SwapUnsafe(origKey, swp)
}

func (c *lookupTTLCache[OK, AK, V]) HasBy(lookup string, key AK) bool {
	c.Lock()
	has := c.lookup.Has(lookup, key)
	c.Unlock()
	return has
}

func (c *lookupTTLCache[OK, AK, V]) InvalidateBy(lookup string, key AK) bool {
	c.Lock()
	defer c.Unlock()
	origKey, ok := c.lookup.Get(lookup, key)
	if !ok {
		return false
	}
	c.InvalidateUnsafe(origKey)
	return true
}

// LookupMap is a structure that provides lookups for
// keys to primary keys under supplied lookup identifiers.
// This is essentially a wrapper around map[string](map[K1]K2).
type LookupMap[OK comparable, AK comparable] struct {
	initd  bool
	lookup map[string](map[AK]OK)
}

// RegisterLookup registers a lookup identifier in the LookupMap,
// note this can only be doing during the cfg.RegisterLookups() hook.
func (l *LookupMap[OK, AK]) RegisterLookup(id string) {
	if l.initd {
		panic("cache: cannot register lookup after initialization")
	} else if _, ok := l.lookup[id]; ok {
		panic("cache: lookup mapping already exists for identifier")
	}
	l.lookup[id] = make(map[AK]OK, 100)
}

// Get fetches an entry's primary key for lookup identifier and key.
func (l *LookupMap[OK, AK]) Get(id string, key AK) (OK, bool) {
	keys, ok := l.lookup[id]
	if !ok {
		var key OK
		return key, false
	}
	origKey, ok := keys[key]
	return origKey, ok
}

// Set adds a lookup to the LookupMap under supplied lookup identifier,
// linking supplied key to the supplied primary (original) key.
func (l *LookupMap[OK, AK]) Set(id string, key AK, origKey OK) {
	keys, ok := l.lookup[id]
	if !ok {
		panic("cache: invalid lookup identifier")
	}
	keys[key] = origKey
}

// Has checks if there exists a lookup for supplied identifier and key.
func (l *LookupMap[OK, AK]) Has(id string, key AK) bool {
	keys, ok := l.lookup[id]
	if !ok {
		return false
	}
	_, ok = keys[key]
	return ok
}

// Delete removes a lookup from LookupMap with supplied identifier and key.
func (l *LookupMap[OK, AK]) Delete(id string, key AK) {
	keys, ok := l.lookup[id]
	if !ok {
		return
	}
	delete(keys, key)
}
