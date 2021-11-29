package cache

import (
	"sync"

	"github.com/ReneKroon/ttlcache"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type TagCache struct {
	cache  *ttlcache.Cache // map of IDS -> cached tags
	mutext sync.Mutex
}

func NewTagCache() *TagCache {
	c := TagCache{
		cache:  ttlcache.NewCache(),
		mutext: sync.Mutex{},
	}

	c.cache.SetExpirationCallback(func(key string, value interface{}) {
		_, ok := value.(*gtsmodel.Tag)
		if !ok {
			logrus.Panicf("TagCache cloud not assert entry with key %s to *gtsmode.Tag", key)
		}

		c.mutext.Lock()
		// do other things here
		c.mutext.Unlock()
	})

	return &c
}

func (c *TagCache) GetByID(id string) (*gtsmodel.Tag, bool) {
	c.mutext.Lock()
	tag, ok := c.getByID(id)
	c.mutext.Unlock()
	return tag, ok
}

func (c *TagCache) getByID(id string) (*gtsmodel.Tag, bool) {
	v, ok := c.cache.Get(id)
	if !ok {
		return nil, false
	}
	return copyTag(v.(*gtsmodel.Tag)), true
}

func copyTag(tag *gtsmodel.Tag) *gtsmodel.Tag {
	return &gtsmodel.Tag{
		ID:                     tag.ID,
		CreatedAt:              tag.CreatedAt,
		UpdatedAt:              tag.UpdatedAt,
		URL:                    tag.URL,
		Name:                   tag.Name,
		FirstSeenFromAccountID: tag.FirstSeenFromAccountID,
		Useable:                tag.Useable,
		Listable:               tag.Listable,
		LastStatusAt:           tag.LastStatusAt,
	}
}
