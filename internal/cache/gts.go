package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type GTSCaches interface {
	// Init will initialize all the gtsmodel caches in this collection.
	// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
	Init()

	// Start will attempt to start all of the gtsmodel caches, or panic.
	Start()

	// Stop will attempt to stop all of the gtsmodel caches, or panic.
	Stop()

	// Account provides access to the gtsmodel Account database cache.
	Account() *result.Cache[*gtsmodel.Account]

	// Block provides access to the gtsmodel Block (account) database cache.
	Block() *result.Cache[*gtsmodel.Block]

	// DomainBlock provides access to the gtsmodel DomainBlock database cache.
	DomainBlock() *result.Cache[*gtsmodel.DomainBlock]

	// Emoji provides access to the gtsmodel Emoji database cache.
	Emoji() *result.Cache[*gtsmodel.Emoji]

	// EmojiCategory provides access to the gtsmodel EmojiCategory database cache.
	EmojiCategory() *result.Cache[*gtsmodel.EmojiCategory]

	// Mention provides access to the gtsmodel Mention database cache.
	Mention() *result.Cache[*gtsmodel.Mention]

	// Notification provides access to the gtsmodel Notification database cache.
	Notification() *result.Cache[*gtsmodel.Notification]

	// Status provides access to the gtsmodel Status database cache.
	Status() *result.Cache[*gtsmodel.Status]

	// Tombstone provides access to the gtsmodel Tombstone database cache.
	Tombstone() *result.Cache[*gtsmodel.Tombstone]

	// User provides access to the gtsmodel User database cache.
	User() *result.Cache[*gtsmodel.User]
}

// NewGTS returns a new default implementation of GTSCaches.
func NewGTS() GTSCaches {
	return &gtsCaches{}
}

type gtsCaches struct {
	account       *result.Cache[*gtsmodel.Account]
	block         *result.Cache[*gtsmodel.Block]
	domainBlock   *result.Cache[*gtsmodel.DomainBlock]
	emoji         *result.Cache[*gtsmodel.Emoji]
	emojiCategory *result.Cache[*gtsmodel.EmojiCategory]
	mention       *result.Cache[*gtsmodel.Mention]
	notification  *result.Cache[*gtsmodel.Notification]
	status        *result.Cache[*gtsmodel.Status]
	tombstone     *result.Cache[*gtsmodel.Tombstone]
	user          *result.Cache[*gtsmodel.User]
}

func (c *gtsCaches) Init() {
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

func (c *gtsCaches) Start() {
	tryUntil("starting gtsmodel.Account cache", 5, func() bool {
		return c.account.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Block cache", 5, func() bool {
		return c.block.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.DomainBlock cache", 5, func() bool {
		return c.domainBlock.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Emoji cache", 5, func() bool {
		return c.emoji.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.EmojiCategory cache", 5, func() bool {
		return c.emojiCategory.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Mention cache", 5, func() bool {
		return c.mention.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Notification cache", 5, func() bool {
		return c.notification.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Status cache", 5, func() bool {
		return c.status.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.Tombstone cache", 5, func() bool {
		return c.tombstone.Start(time.Second * 10)
	})
	tryUntil("starting gtsmodel.User cache", 5, func() bool {
		return c.user.Start(time.Second * 10)
	})
}

func (c *gtsCaches) Stop() {
	tryUntil("stopping gtsmodel.Account cache", 5, c.account.Stop)
	tryUntil("stopping gtsmodel.Block cache", 5, c.block.Stop)
	tryUntil("stopping gtsmodel.DomainBlock cache", 5, c.domainBlock.Stop)
	tryUntil("stopping gtsmodel.Emoji cache", 5, c.emoji.Stop)
	tryUntil("stopping gtsmodel.EmojiCategory cache", 5, c.emojiCategory.Stop)
	tryUntil("stopping gtsmodel.Mention cache", 5, c.mention.Stop)
	tryUntil("stopping gtsmodel.Notification cache", 5, c.notification.Stop)
	tryUntil("stopping gtsmodel.Status cache", 5, c.status.Stop)
	tryUntil("stopping gtsmodel.Tombstone cache", 5, c.tombstone.Stop)
	tryUntil("stopping gtsmodel.User cache", 5, c.user.Stop)
}

func (c *gtsCaches) Account() *result.Cache[*gtsmodel.Account] {
	return c.account
}

func (c *gtsCaches) Block() *result.Cache[*gtsmodel.Block] {
	return c.block
}

func (c *gtsCaches) DomainBlock() *result.Cache[*gtsmodel.DomainBlock] {
	return c.domainBlock
}

func (c *gtsCaches) Emoji() *result.Cache[*gtsmodel.Emoji] {
	return c.emoji
}

func (c *gtsCaches) EmojiCategory() *result.Cache[*gtsmodel.EmojiCategory] {
	return c.emojiCategory
}

func (c *gtsCaches) Mention() *result.Cache[*gtsmodel.Mention] {
	return c.mention
}

func (c *gtsCaches) Notification() *result.Cache[*gtsmodel.Notification] {
	return c.notification
}

func (c *gtsCaches) Status() *result.Cache[*gtsmodel.Status] {
	return c.status
}

func (c *gtsCaches) Tombstone() *result.Cache[*gtsmodel.Tombstone] {
	return c.tombstone
}

func (c *gtsCaches) User() *result.Cache[*gtsmodel.User] {
	return c.user
}

func (c *gtsCaches) initAccount() {
	c.account = result.NewSized([]result.Lookup{
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
	c.account.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initBlock() {
	c.block = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "URI"},
	}, func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1
		return b2
	}, 1000)
	c.block.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initDomainBlock() {
	c.domainBlock = result.NewSized([]result.Lookup{
		{Name: "Domain"},
	}, func(d1 *gtsmodel.DomainBlock) *gtsmodel.DomainBlock {
		d2 := new(gtsmodel.DomainBlock)
		*d2 = *d1
		return d2
	}, 1000)
	c.domainBlock.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initEmoji() {
	c.emoji = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain"},
		{Name: "ImageStaticURL"},
	}, func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1
		return e2
	}, 1000)
	c.emoji.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initEmojiCategory() {
	c.emojiCategory = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}, 1000)
	c.emojiCategory.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initMention() {
	c.mention = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1
		return m2
	}, 1000)
	c.mention.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initNotification() {
	c.notification = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1
		return n2
	}, 1000)
	c.notification.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initStatus() {
	c.status = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
	}, func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1
		return s2
	}, 1000)
	c.status.SetTTL(time.Minute*5, false)
}

// initTombstone will initialize the gtsmodel.Tombstone cache.
func (c *gtsCaches) initTombstone() {
	c.tombstone = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, 100)
	c.tombstone.SetTTL(time.Minute*5, false)
}

func (c *gtsCaches) initUser() {
	c.user = result.NewSized([]result.Lookup{
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
	c.user.SetTTL(time.Minute*5, false)
}
