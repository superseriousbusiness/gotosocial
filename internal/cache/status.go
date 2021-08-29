package cache

import (
	"sync"

	"github.com/ReneKroon/ttlcache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// statusCache is a wrapper around ttlcache.Cache to provide URL and URI lookups for gtsmodel.Status
type StatusCache struct {
	cache *ttlcache.Cache   // map of IDs -> cached statuses
	urls  map[string]string // map of status URLs -> IDs
	uris  map[string]string // map of status URIs -> IDs
	mutex sync.Mutex
}

// newStatusCache returns a new instantiated statusCache object
func NewStatusCache() *StatusCache {
	c := StatusCache{
		cache: ttlcache.NewCache(),
		urls:  make(map[string]string, 100),
		uris:  make(map[string]string, 100),
		mutex: sync.Mutex{},
	}

	// Set callback to purge lookup maps on expiration
	c.cache.SetExpirationCallback(func(key string, value interface{}) {
		status := value.(*gtsmodel.Status)

		c.mutex.Lock()
		delete(c.urls, status.URL)
		delete(c.uris, status.URI)
		c.mutex.Unlock()
	})

	return &c
}

// GetByID attempts to fetch a status from the cache by its ID
func (c *StatusCache) GetByID(id string) (*gtsmodel.Status, bool) {
	c.mutex.Lock()
	status, ok := c.getByID(id)
	c.mutex.Unlock()
	return status, ok
}

// GetByURL attempts to fetch a status from the cache by its URL
func (c *StatusCache) GetByURL(url string) (*gtsmodel.Status, bool) {
	// Perform safe ID lookup
	c.mutex.Lock()
	id, ok := c.urls[url]

	// Not found, unlock early
	if !ok {
		c.mutex.Unlock()
		return nil, false
	}

	// Attempt status lookup
	status, ok := c.getByID(id)
	c.mutex.Unlock()
	return status, ok
}

// GetByURI attempts to fetch a status from the cache by its URI
func (c *StatusCache) GetByURI(uri string) (*gtsmodel.Status, bool) {
	// Perform safe ID lookup
	c.mutex.Lock()
	id, ok := c.uris[uri]

	// Not found, unlock early
	if !ok {
		c.mutex.Unlock()
		return nil, false
	}

	// Attempt status lookup
	status, ok := c.getByID(id)
	c.mutex.Unlock()
	return status, ok
}

// getByID performs an unsafe (no mutex locks) lookup of status by ID
func (c *StatusCache) getByID(id string) (*gtsmodel.Status, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return v.(*gtsmodel.Status), true
}

// Put places a status in the cache
func (c *StatusCache) Put(status *gtsmodel.Status) {
	if status == nil || status.ID == "" ||
		status.URL == "" ||
		status.URI == "" {
		panic("invalid status")
	}

	c.mutex.Lock()
	c.cache.Set(status.ID, status)
	c.urls[status.URL] = status.ID
	c.uris[status.URI] = status.ID
	c.mutex.Unlock()
}
