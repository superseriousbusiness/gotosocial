package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v3/result"
)

type VisibilityCache struct {
	*result.Cache[*CachedVisibility]
}

func (c *VisibilityCache) Init() {
	c.Cache = result.New([]result.Lookup{
		{Name: "ItemID"},
		{Name: "RequesterID"},
		{Name: "Type.RequesterID.ItemID"},
	}, func(v1 *CachedVisibility) *CachedVisibility {
		v2 := new(CachedVisibility)
		*v2 = *v1
		return v2
	}, 1000)
	c.Cache.SetTTL(time.Minute*30, true)
}

func (c *VisibilityCache) Start() {
	tryUntil("starting visibility cache", 5, func() bool {
		return c.Cache.Start(time.Minute)
	})
}

func (c *VisibilityCache) Stop() {
	tryUntil("stopping visibility cache", 5, c.Cache.Stop)
}

// CachedVisibility represents a cached visibility lookup value.
type CachedVisibility struct {
	// ItemID is the ID of the item in question (status / account).
	ItemID string

	// RequesterID is the ID of the requesting account for this visibility lookup.
	RequesterID string

	// Type is the visibility lookup type: ["status", "account", "home", "public"]
	Type string

	// Value is the actual visibility value.
	Value bool
}
