package cache

// LookupCfg is the LookupCache configuration
type LookupCfg struct {
	// RegisterLookups is called on init to register lookups
	// within LookupCache's internal LookupMap
	RegisterLookups func(*LookupMap)

	// AddLookups is called on each addition to the cache, to
	// set any required additional key lookups for supplied item
	AddLookups func(*LookupMap, interface{})

	// DeleteLookups is called on each eviction/invalidation of
	// an item in the cache, to remove any unused key lookups
	DeleteLookups func(*LookupMap, interface{})
}

// LookupCache is a cache built on-top of TTLCache, providing multi-key
// lookups for items in the cache by means of additional lookup maps. These
// maps simply store addtional keys => original key, with hook-ins to automatically
// call user supplied functions on adding an item, or on updating/deleting an
// item to keep the LookupMap up-to-date
type LookupCache interface {
	Cache

	// GetBy fetches a cached value by supplied lookup identifier and key
	GetBy(lookup string, key string) (interface{}, bool)

	// HasBy checks if a value is cached under supplied lookup identifier and key
	HasBy(lookup string, key string) bool

	// InvalidateBy invalidates a value by supplied lookup identifier and key
	InvalidateBy(lookup string, key string) bool
}

type lookupTTLCache struct {
	config LookupCfg
	lookup LookupMap
	TTLCache
}

// NewLookup returns a new initialized LookupCache
func NewLookup(cfg LookupCfg) LookupCache {
	switch {
	case cfg.RegisterLookups == nil:
		panic("cache: nil lookups register function")
	case cfg.AddLookups == nil:
		panic("cache: nil lookups add function")
	case cfg.DeleteLookups == nil:
		panic("cache: nil delete lookups function")
	}
	c := lookupTTLCache{config: cfg}
	c.TTLCache.Init()
	c.lookup.lookup = map[string]map[string]string{}
	c.config.RegisterLookups(&c.lookup)
	c.SetEvictionCallback(nil)
	c.SetInvalidateCallback(nil)
	c.lookup.initd = true
	return &c
}

func (c *lookupTTLCache) SetEvictionCallback(hook Hook) {
	if hook == nil {
		hook = emptyHook
	}
	c.TTLCache.SetEvictionCallback(func(key string, value interface{}) {
		hook(key, value)
		c.config.DeleteLookups(&c.lookup, value)
	})
}

func (c *lookupTTLCache) SetInvalidateCallback(hook Hook) {
	if hook == nil {
		hook = emptyHook
	}
	c.TTLCache.SetInvalidateCallback(func(key string, value interface{}) {
		hook(key, value)
		c.config.DeleteLookups(&c.lookup, value)
	})
}

func (c *lookupTTLCache) GetBy(lookup string, key string) (interface{}, bool) {
	c.Lock()
	origKey, ok := c.lookup.Get(lookup, key)
	if !ok {
		c.Unlock()
		return nil, false
	}
	v, ok := c.GetUnsafe(origKey)
	c.Unlock()
	return v, ok
}

func (c *lookupTTLCache) Put(key string, value interface{}) bool {
	c.Lock()
	put := c.PutUnsafe(key, value)
	if put {
		c.config.AddLookups(&c.lookup, value)
	}
	c.Unlock()
	return put
}

func (c *lookupTTLCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	c.SetUnsafe(key, value)
	c.config.AddLookups(&c.lookup, value)
}

func (c *lookupTTLCache) HasBy(lookup string, key string) bool {
	c.Lock()
	has := c.lookup.Has(lookup, key)
	c.Unlock()
	return has
}

func (c *lookupTTLCache) InvalidateBy(lookup string, key string) bool {
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
// This is essentially a wrapper around map[string](map[string]string)
type LookupMap struct {
	initd  bool
	lookup map[string](map[string]string)
}

// RegisterLookup registers a lookup identifier in the LookupMap,
// note this can only be doing during the cfg.RegisterLookups() hook
func (l *LookupMap) RegisterLookup(id string) {
	if l.initd {
		panic("cache: cannot register lookup after initialization")
	} else if _, ok := l.lookup[id]; ok {
		panic("cache: lookup mapping already exists for identifier")
	}
	l.lookup[id] = make(map[string]string, 100)
}

// Get fetches an entry's primary key for lookup identifier and key
func (l *LookupMap) Get(id string, key string) (string, bool) {
	keys, ok := l.lookup[id]
	if !ok {
		return "", false
	}
	origKey, ok := keys[key]
	return origKey, ok
}

// Set adds a lookup to the LookupMap under supplied lookup identifer,
// linking supplied key to the supplied primary (original) key
func (l *LookupMap) Set(id string, key string, origKey string) {
	keys, ok := l.lookup[id]
	if !ok {
		panic("cache: invalid lookup identifier")
	}
	keys[key] = origKey
}

// Has checks if there exists a lookup for supplied identifer and key
func (l *LookupMap) Has(id string, key string) bool {
	keys, ok := l.lookup[id]
	if !ok {
		return false
	}
	_, ok = keys[key]
	return ok
}

// Delete removes a lookup from LookupMap with supplied identifer and key
func (l *LookupMap) Delete(id string, key string) {
	keys, ok := l.lookup[id]
	if !ok {
		return
	}
	delete(keys, key)
}
