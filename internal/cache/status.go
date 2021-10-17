package cache

import (
	"codeberg.org/gruf/go-cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// StatusCache is a wrapper around ttlcache.Cache to provide URL and URI lookups for gtsmodel.Status
type StatusCache struct {
	cache cache.LookupCache
}

// NewStatusCache returns a new instantiated statusCache object
func NewStatusCache() *StatusCache {
	cache := cache.NewLookup(cache.LookupCfg{
		RegisterLookups: func(lm *cache.LookupMap) {
			lm.RegisterLookup("url")
			lm.RegisterLookup("uri")
		},

		AddLookups: func(lm *cache.LookupMap, i interface{}) {
			st := i.(*gtsmodel.Status)
			lm.Set("uri", st.URI, st.ID)
			lmSetIf(lm, "url", st.URL, st.ID)
		},

		DeleteLookups: func(lm *cache.LookupMap, i interface{}) {
			st := i.(*gtsmodel.Status)
			lm.Delete("uri", st.URI)
			lmDeleteIf(lm, "url", st.URL)
		},
	})
	return &StatusCache{cache: cache}
}

// GetByID attempts to fetch a status from the cache by its ID, you will receive a copy for thread-safety
func (c *StatusCache) GetByID(id string) (*gtsmodel.Status, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return copyStatus(v.(*gtsmodel.Status)), true
}

// GetByURL attempts to fetch a status from the cache by its URL, you will receive a copy for thread-safety
func (c *StatusCache) GetByURL(url string) (*gtsmodel.Status, bool) {
	v, ok := c.cache.GetBy("url", url)
	if !ok {
		return nil, false
	}
	return copyStatus(v.(*gtsmodel.Status)), true
}

// GetByURI attempts to fetch a status from the cache by its URI, you will receive a copy for thread-safety
func (c *StatusCache) GetByURI(uri string) (*gtsmodel.Status, bool) {
	v, ok := c.cache.GetBy("uri", uri)
	if !ok {
		return nil, false
	}
	return copyStatus(v.(*gtsmodel.Status)), true
}

// getByID performs an unsafe (no mutex locks) lookup of status by ID, returning a copy of status in cache
func (c *StatusCache) getByID(id string) (*gtsmodel.Status, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return copyStatus(v.(*gtsmodel.Status)), true
}

// Put places a status in the cache, ensuring that the object place is a copy for thread-safety
func (c *StatusCache) Put(status *gtsmodel.Status) {
	if status == nil || status.ID == "" || status.URI == "" {
		panic("invalid status")
	}
	c.cache.Set(status.ID, copyStatus(status))
}

// copyStatus performs a surface-level copy of status, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyStatus(status *gtsmodel.Status) *gtsmodel.Status {
	st := *status
	st.Attachments = nil
	st.Tags = nil
	st.Mentions = nil
	st.Emojis = nil
	st.Account = nil
	st.InReplyTo = nil
	st.InReplyToAccount = nil
	st.BoostOf = nil
	st.BoostOfAccount = nil
	return &st
}
