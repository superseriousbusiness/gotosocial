package cache

import (
	"sync"

	"github.com/ReneKroon/ttlcache"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// AccountCache is a wrapper around ttlcache.Cache to provide URL and URI lookups for gtsmodel.Account
type AccountCache struct {
	cache *ttlcache.Cache   // map of IDs -> cached accounts
	urls  map[string]string // map of account URLs -> IDs
	uris  map[string]string // map of account URIs -> IDs
	mutex sync.Mutex
}

// NewAccountCache returns a new instantiated AccountCache object
func NewAccountCache() *AccountCache {
	c := AccountCache{
		cache: ttlcache.NewCache(),
		urls:  make(map[string]string, 100),
		uris:  make(map[string]string, 100),
		mutex: sync.Mutex{},
	}

	// Set callback to purge lookup maps on expiration
	c.cache.SetExpirationCallback(func(key string, value interface{}) {
		account, ok := value.(*gtsmodel.Account)
		if !ok {
			logrus.Panicf("AccountCache could not assert entry with key %s to *gtsmodel.Account", key)
		}

		c.mutex.Lock()
		delete(c.urls, account.URL)
		delete(c.uris, account.URI)
		c.mutex.Unlock()
	})

	return &c
}

// GetByID attempts to fetch a account from the cache by its ID, you will receive a copy for thread-safety
func (c *AccountCache) GetByID(id string) (*gtsmodel.Account, bool) {
	c.mutex.Lock()
	account, ok := c.getByID(id)
	c.mutex.Unlock()
	return account, ok
}

// GetByURL attempts to fetch a account from the cache by its URL, you will receive a copy for thread-safety
func (c *AccountCache) GetByURL(url string) (*gtsmodel.Account, bool) {
	// Perform safe ID lookup
	c.mutex.Lock()
	id, ok := c.urls[url]

	// Not found, unlock early
	if !ok {
		c.mutex.Unlock()
		return nil, false
	}

	// Attempt account lookup
	account, ok := c.getByID(id)
	c.mutex.Unlock()
	return account, ok
}

// GetByURI attempts to fetch a account from the cache by its URI, you will receive a copy for thread-safety
func (c *AccountCache) GetByURI(uri string) (*gtsmodel.Account, bool) {
	// Perform safe ID lookup
	c.mutex.Lock()
	id, ok := c.uris[uri]

	// Not found, unlock early
	if !ok {
		c.mutex.Unlock()
		return nil, false
	}

	// Attempt account lookup
	account, ok := c.getByID(id)
	c.mutex.Unlock()
	return account, ok
}

// getByID performs an unsafe (no mutex locks) lookup of account by ID, returning a copy of account in cache
func (c *AccountCache) getByID(id string) (*gtsmodel.Account, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}

	a, ok := v.(*gtsmodel.Account)
	if !ok {
		panic("account cache entry was not an account")
	}

	return copyAccount(a), true
}

// Put places a account in the cache, ensuring that the object place is a copy for thread-safety
func (c *AccountCache) Put(account *gtsmodel.Account) {
	if account == nil || account.ID == "" {
		panic("invalid account")
	}

	c.mutex.Lock()
	c.cache.Set(account.ID, copyAccount(account))
	if account.URL != "" {
		c.urls[account.URL] = account.ID
	}
	if account.URI != "" {
		c.uris[account.URI] = account.ID
	}
	c.mutex.Unlock()
}

// copyAccount performs a surface-level copy of account, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyAccount(account *gtsmodel.Account) *gtsmodel.Account {
	return &gtsmodel.Account{
		ID:                      account.ID,
		Username:                account.Username,
		Domain:                  account.Domain,
		AvatarMediaAttachmentID: account.AvatarMediaAttachmentID,
		AvatarMediaAttachment:   nil,
		AvatarRemoteURL:         account.AvatarRemoteURL,
		HeaderMediaAttachmentID: account.HeaderMediaAttachmentID,
		HeaderMediaAttachment:   nil,
		HeaderRemoteURL:         account.HeaderRemoteURL,
		DisplayName:             account.DisplayName,
		Fields:                  account.Fields,
		Note:                    account.Note,
		NoteRaw:                 account.NoteRaw,
		Memorial:                account.Memorial,
		MovedToAccountID:        account.MovedToAccountID,
		CreatedAt:               account.CreatedAt,
		UpdatedAt:               account.UpdatedAt,
		Bot:                     account.Bot,
		Reason:                  account.Reason,
		Locked:                  account.Locked,
		Discoverable:            account.Discoverable,
		Privacy:                 account.Privacy,
		Sensitive:               account.Sensitive,
		Language:                account.Language,
		URI:                     account.URI,
		URL:                     account.URL,
		LastWebfingeredAt:       account.LastWebfingeredAt,
		InboxURI:                account.InboxURI,
		OutboxURI:               account.OutboxURI,
		FollowingURI:            account.FollowingURI,
		FollowersURI:            account.FollowersURI,
		FeaturedCollectionURI:   account.FeaturedCollectionURI,
		ActorType:               account.ActorType,
		AlsoKnownAs:             account.AlsoKnownAs,
		PrivateKey:              account.PrivateKey,
		PublicKey:               account.PublicKey,
		PublicKeyURI:            account.PublicKeyURI,
		SensitizedAt:            account.SensitizedAt,
		SilencedAt:              account.SilencedAt,
		SuspendedAt:             account.SuspendedAt,
		HideCollections:         account.HideCollections,
		SuspensionOrigin:        account.SuspensionOrigin,
	}
}
