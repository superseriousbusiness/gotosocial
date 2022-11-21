package cache

type APCaches interface {
	// Init will initialize all the ActivityPub caches in this collection.
	// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
	Init()

	// Start will attempt to start all of the ActivityPub caches, or panic.
	Start()

	// Stop will attempt to stop all of the ActivityPub caches, or panic.
	Stop()
}

// NewAP returns a new default implementation of APCaches.
func NewAP() APCaches {
	return &apCaches{}
}

type apCaches struct{}

func (c *apCaches) Init() {}

func (c *apCaches) Start() {}

func (c *apCaches) Stop() {}
