package cache

type Caches struct {
	// GTS provides access to the collection of gtsmodel object caches.
	GTS GTSCaches

	// AP provides access to the collection of ActivityPub object caches.
	AP APCaches

	// prevent pass-by-value.
	_ nocopy
}

// Init will (re)initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
	if c.GTS == nil {
		// use default impl
		c.GTS = NewGTS()
	}

	if c.AP == nil {
		// use default impl
		c.AP = NewAP()
	}

	// initialize caches
	c.GTS.Init()
	c.AP.Init()
}

// Start will start both the GTS and AP cache collections.
func (c *Caches) Start() {
	c.GTS.Start()
	c.AP.Start()
}

// Stop will stop both the GTS and AP cache collections.
func (c *Caches) Stop() {
	c.GTS.Stop()
	c.AP.Stop()
}
