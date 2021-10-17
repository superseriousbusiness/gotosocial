package cache

import (
	"codeberg.org/gruf/go-cache"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// AccountCache is a wrapper around ttlcache.Cache to provide URL and URI lookups for gtsmodel.Account
type AccountCache struct {
	cache cache.LookupCache // map of IDs -> cached accounts
}

// NewAccountCache returns a new instantiated AccountCache object
func NewAccountCache() *AccountCache {
	cache := cache.NewLookup(cache.LookupCfg{
		RegisterLookups: func(lm *cache.LookupMap) {
			lm.RegisterLookup("url")
			lm.RegisterLookup("uri")
		},

		AddLookups: func(lm *cache.LookupMap, i interface{}) {
			acc := i.(*gtsmodel.Account)
			lm.Set("uri", acc.URI, acc.ID)
			lmSetIf(lm, "url", acc.URI, acc.ID)
		},

		DeleteLookups: func(lm *cache.LookupMap, i interface{}) {
			acc := i.(*gtsmodel.Account)
			lm.Delete("uri", acc.URI)
			lmDeleteIf(lm, "url", acc.URL)
		},
	})
	return &AccountCache{cache: cache}
}

// GetByID attempts to fetch a account from the cache by its ID, you will receive a copy for thread-safety
func (c *AccountCache) GetByID(id string) (*gtsmodel.Account, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return copyAccount(v.(*gtsmodel.Account)), true
}

// GetByURL attempts to fetch a account from the cache by its URL, you will receive a copy for thread-safety
func (c *AccountCache) GetByURL(url string) (*gtsmodel.Account, bool) {
	v, ok := c.cache.GetBy("URL", url)
	if !ok {
		return nil, false
	}
	return copyAccount(v.(*gtsmodel.Account)), true
}

// GetByURI attempts to fetch a account from the cache by its URI, you will receive a copy for thread-safety
func (c *AccountCache) GetByURI(uri string) (*gtsmodel.Account, bool) {
	v, ok := c.cache.GetBy("uri", uri)
	if !ok {
		return nil, false
	}
	return copyAccount(v.(*gtsmodel.Account)), true
}

// getByID performs an unsafe (no mutex locks) lookup of account by ID, returning a copy of account in cache
func (c *AccountCache) getByID(id string) (*gtsmodel.Account, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return copyAccount(v.(*gtsmodel.Account)), true
}

// Put places a account in the cache, ensuring that the object place is a copy for thread-safety
func (c *AccountCache) Put(account *gtsmodel.Account) {
	if account == nil || account.ID == "" || account.URI == "" {
		panic("invalid account")
	}
	c.cache.Set(account.ID, copyAccount(account))
}

// copyAccount performs a surface-level copy of account, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyAccount(account *gtsmodel.Account) *gtsmodel.Account {
	acc := *account
	acc.AvatarMediaAttachment = nil
	acc.HeaderMediaAttachment = nil
	return &acc
}
