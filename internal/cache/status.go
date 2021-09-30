package cache

import (
	"sync"

	"github.com/ReneKroon/ttlcache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// StatusCache is a wrapper around ttlcache.Cache to provide URL and URI lookups for gtsmodel.Status
type StatusCache struct {
	cache *ttlcache.Cache   // map of IDs -> cached statuses
	urls  map[string]string // map of status URLs -> IDs
	uris  map[string]string // map of status URIs -> IDs
	mutex sync.Mutex
}

// NewStatusCache returns a new instantiated statusCache object
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

// GetByID attempts to fetch a status from the cache by its ID, you will receive a copy for thread-safety
func (c *StatusCache) GetByID(id string) (*gtsmodel.Status, bool) {
	c.mutex.Lock()
	status, ok := c.getByID(id)
	c.mutex.Unlock()
	return status, ok
}

// GetByURL attempts to fetch a status from the cache by its URL, you will receive a copy for thread-safety
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

// GetByURI attempts to fetch a status from the cache by its URI, you will receive a copy for thread-safety
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
	if status == nil || status.ID == "" {
		panic("invalid status")
	}

	c.mutex.Lock()
	c.cache.Set(status.ID, copyStatus(status))
	if status.URL != "" {
		c.urls[status.URL] = status.ID
	}
	if status.URI != "" {
		c.uris[status.URI] = status.ID
	}
	c.mutex.Unlock()
}

// copyStatus performs a surface-level copy of status, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyStatus(status *gtsmodel.Status) *gtsmodel.Status {
	return &gtsmodel.Status{
		ID:                       status.ID,
		URI:                      status.URI,
		URL:                      status.URL,
		Content:                  status.Content,
		AttachmentIDs:            status.AttachmentIDs,
		Attachments:              nil,
		TagIDs:                   status.TagIDs,
		Tags:                     nil,
		MentionIDs:               status.MentionIDs,
		Mentions:                 nil,
		EmojiIDs:                 status.EmojiIDs,
		Emojis:                   nil,
		CreatedAt:                status.CreatedAt,
		UpdatedAt:                status.UpdatedAt,
		Local:                    status.Local,
		AccountID:                status.AccountID,
		Account:                  nil,
		AccountURI:               status.AccountURI,
		InReplyToID:              status.InReplyToID,
		InReplyTo:                nil,
		InReplyToURI:             status.InReplyToURI,
		InReplyToAccountID:       status.InReplyToAccountID,
		InReplyToAccount:         nil,
		BoostOfID:                status.BoostOfID,
		BoostOf:                  nil,
		BoostOfAccountID:         status.BoostOfAccountID,
		BoostOfAccount:           nil,
		ContentWarning:           status.ContentWarning,
		Visibility:               status.Visibility,
		Sensitive:                status.Sensitive,
		Language:                 status.Language,
		CreatedWithApplicationID: status.CreatedWithApplicationID,
		Federated:                status.Federated,
		Boostable:                status.Boostable,
		Replyable:                status.Replyable,
		Likeable:                 status.Likeable,
		ActivityStreamsType:      status.ActivityStreamsType,
		Text:                     status.Text,
		Pinned:                   status.Pinned,
	}
}
