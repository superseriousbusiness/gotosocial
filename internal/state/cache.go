package state

import (
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

	"codeberg.org/gruf/go-cache/v3/result"
)

type Caches struct {
	// GTS provides access to the collection of gtsmodel object caches.
	GTS GTSCaches

	// AP provides access to the collection of ActivityPub object caches.
	AP APCaches

	// prevent pass-by-value.
	_ nocopy
}

// Init will initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
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

type GTSCaches struct {
	Account       *result.Cache[*gtsmodel.Account]
	Block         *result.Cache[*gtsmodel.Block]
	DomainBlock   *result.Cache[*gtsmodel.DomainBlock]
	Emoji         *result.Cache[*gtsmodel.Emoji]
	EmojiCategory *result.Cache[*gtsmodel.EmojiCategory]
	Mention       *result.Cache[*gtsmodel.Mention]
	Notification  *result.Cache[*gtsmodel.Notification]
	Status        *result.Cache[*gtsmodel.Status]
	Tombstone     *result.Cache[*gtsmodel.Tombstone]
	User          *result.Cache[*gtsmodel.User]
}

// Init will initialize all the gtsmodel caches in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *GTSCaches) Init() {
	c.initAccount()
	c.initBlock()
	c.initDomainBlock()
	c.initEmoji()
	c.initEmojiCategory()
	c.initMention()
	c.initNotification()
	c.initStatus()
	c.initTombstone()
	c.initUser()
}

// Start will attempt to start all of the gtsmodel caches, or panic.
func (c *GTSCaches) Start() {
	tryUntil("starting gtsmodel.Account cache", 5, func() bool {
		return c.Account.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Block cache", 5, func() bool {
		return c.Block.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.DomainBlock cache", 5, func() bool {
		return c.DomainBlock.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Emoji cache", 5, func() bool {
		return c.Emoji.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.EmojiCategory cache", 5, func() bool {
		return c.EmojiCategory.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Mention cache", 5, func() bool {
		return c.Mention.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Notification cache", 5, func() bool {
		return c.Notification.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Status cache", 5, func() bool {
		return c.Status.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Tombstone cache", 5, func() bool {
		return c.Tombstone.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.User cache", 5, func() bool {
		return c.User.Start(time.Second * 10)
	})
}

// Stop will attempt to stop all of the gtsmodel caches, or panic.
func (c *GTSCaches) Stop() {
	tryUntil("stopping gtsmodel.Account cache", 5, c.Account.Stop)
	tryUntil("stopping gtsmodel.Block cache", 5, c.Block.Stop)
	tryUntil("stopping gtsmodel.DomainBlock cache", 5, c.DomainBlock.Stop)
	tryUntil("stopping gtsmodel.Emoji cache", 5, c.Emoji.Stop)
	tryUntil("stopping gtsmodel.EmojiCategory cache", 5, c.EmojiCategory.Stop)
	tryUntil("stopping gtsmodel.Mention cache", 5, c.Mention.Stop)
	tryUntil("stopping gtsmodel.Notification cache", 5, c.Notification.Stop)
	tryUntil("stopping gtsmodel.Status cache", 5, c.Status.Stop)
	tryUntil("stopping gtsmodel.Tombstone cache", 5, c.Tombstone.Stop)
	tryUntil("stopping gtsmodel.User cache", 5, c.User.Stop)
}

// initAccount will initialize the gtsmodel.Account cache.
func (c *GTSCaches) initAccount() {
	c.Account = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "Username.Domain"},
		{Name: "PublicKeyURI"},
	}, func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1
		return a2
	}, 1000)
	c.Account.SetTTL(time.Minute*5, false)
}

// initBlock will initialize the gtsmodel.Block cache.
func (c *GTSCaches) initBlock() {
	c.Block = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "URI"},
	}, func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1
		return b2
	}, 1000)
	c.DomainBlock.SetTTL(time.Minute*5, false)
}

// initDomainBlock will initialize the gtsmodel.DomainBlock cache.
func (c *GTSCaches) initDomainBlock() {
	c.DomainBlock = result.NewSized([]result.Lookup{
		{Name: "Domain"},
	}, func(d1 *gtsmodel.DomainBlock) *gtsmodel.DomainBlock {
		d2 := new(gtsmodel.DomainBlock)
		*d2 = *d1
		return d2
	}, 1000)
	c.DomainBlock.SetTTL(time.Minute*5, false)
}

// initEmoji will initialize the gtsmodel.Emoji cache.
func (c *GTSCaches) initEmoji() {
	c.Emoji = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain"},
		{Name: "ImageStaticURL"},
	}, func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1
		return e2
	}, 1000)
	c.Emoji.SetTTL(time.Minute*5, false)
}

// initEmojiCategory will initialize the gtsmodel.EmojiCategory cache.
func (c *GTSCaches) initEmojiCategory() {
	c.EmojiCategory = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}, 1000)
	c.EmojiCategory.SetTTL(time.Minute*5, false)
}

// initMention will initialize the gtsmodel.Mention cache.
func (c *GTSCaches) initMention() {
	c.Mention = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1
		return m2
	}, 1000)
	c.Mention.SetTTL(time.Minute*5, false)
}

// initNotification will initialize the gtsmodel.Notification cache.
func (c *GTSCaches) initNotification() {
	c.Notification = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1
		return n2
	}, 1000)
	c.Notification.SetTTL(time.Minute*5, false)
}

// initStatus will initialize the gtsmodel.Status cache.
func (c *GTSCaches) initStatus() {
	c.Status = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
	}, func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1
		return s2
	}, 1000)
	c.Status.SetTTL(time.Minute*5, false)
}

// initTombstone will initialize the gtsmodel.Tombstone cache.
func (c *GTSCaches) initTombstone() {
	c.Tombstone = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, 100)
	c.Tombstone.SetTTL(time.Minute*5, false)
}

// initUser will initialize the gtsmodel.User cache.
func (c *GTSCaches) initUser() {
	c.User = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID"},
		{Name: "Email"},
		{Name: "ConfirmationToken"},
		{Name: "ExternalID"},
	}, func(u1 *gtsmodel.User) *gtsmodel.User {
		u2 := new(gtsmodel.User)
		*u2 = *u1
		return u2
	}, 1000)
	c.User.SetTTL(time.Minute*5, false)
}

type APCaches struct{}

// Init will initialize all the ActivityPub caches in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *APCaches) Init() {}

// Start will attempt to start all of the ActivityPub caches, or panic.
func (c *APCaches) Start() {}

// Stop will attempt to stop all of the ActivityPub caches, or panic.
func (c *APCaches) Stop() {}
