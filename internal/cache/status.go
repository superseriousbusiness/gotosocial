package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// StatusCache is a cache wrapper to provide URL and URI lookups for gtsmodel.Status
type StatusCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.Status]
}

// NewStatusCache returns a new instantiated statusCache object
func NewStatusCache() *StatusCache {
	c := &StatusCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.Status]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("uri")
			lm.RegisterLookup("url")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], status *gtsmodel.Status) {
			if uri := status.URI; uri != "" {
				lm.Set("uri", uri, status.ID)
			}
			if url := status.URL; url != "" {
				lm.Set("url", url, status.ID)
			}
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], status *gtsmodel.Status) {
			if uri := status.URI; uri != "" {
				lm.Delete("uri", uri)
			}
			if url := status.URL; url != "" {
				lm.Delete("url", url)
			}
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch a status from the cache by its ID, you will receive a copy for thread-safety
func (c *StatusCache) GetByID(id string) (*gtsmodel.Status, bool) {
	return c.cache.Get(id)
}

// GetByURL attempts to fetch a status from the cache by its URL, you will receive a copy for thread-safety
func (c *StatusCache) GetByURL(url string) (*gtsmodel.Status, bool) {
	return c.cache.GetBy("url", url)
}

// GetByURI attempts to fetch a status from the cache by its URI, you will receive a copy for thread-safety
func (c *StatusCache) GetByURI(uri string) (*gtsmodel.Status, bool) {
	return c.cache.GetBy("uri", uri)
}

// Put places a status in the cache, ensuring that the object place is a copy for thread-safety
func (c *StatusCache) Put(status *gtsmodel.Status) {
	if status == nil || status.ID == "" {
		panic("invalid status")
	}
	c.cache.Set(status.ID, copyStatus(status))
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
