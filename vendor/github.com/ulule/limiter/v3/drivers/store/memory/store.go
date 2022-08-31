package memory

import (
	"context"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/common"
	"github.com/ulule/limiter/v3/internal/bytebuffer"
)

// Store is the in-memory store.
type Store struct {
	// Prefix used for the key.
	Prefix string
	// cache used to store values in-memory.
	cache *CacheWrapper
}

// NewStore creates a new instance of memory store with defaults.
func NewStore() limiter.Store {
	return NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
	})
}

// NewStoreWithOptions creates a new instance of memory store with options.
func NewStoreWithOptions(options limiter.StoreOptions) limiter.Store {
	return &Store{
		Prefix: options.Prefix,
		cache:  NewCache(options.CleanUpInterval),
	}
}

// Get returns the limit for given identifier.
func (store *Store) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	buffer := bytebuffer.New()
	defer buffer.Close()
	buffer.Concat(store.Prefix, ":", key)

	count, expiration := store.cache.Increment(buffer.String(), 1, rate.Period)

	lctx := common.GetContextFromState(time.Now(), rate, expiration, count)
	return lctx, nil
}

// Increment increments the limit by given count & returns the new limit value for given identifier.
func (store *Store) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	buffer := bytebuffer.New()
	defer buffer.Close()
	buffer.Concat(store.Prefix, ":", key)

	newCount, expiration := store.cache.Increment(buffer.String(), count, rate.Period)

	lctx := common.GetContextFromState(time.Now(), rate, expiration, newCount)
	return lctx, nil
}

// Peek returns the limit for given identifier, without modification on current values.
func (store *Store) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	buffer := bytebuffer.New()
	defer buffer.Close()
	buffer.Concat(store.Prefix, ":", key)

	count, expiration := store.cache.Get(buffer.String(), rate.Period)

	lctx := common.GetContextFromState(time.Now(), rate, expiration, count)
	return lctx, nil
}

// Reset returns the limit for given identifier.
func (store *Store) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	buffer := bytebuffer.New()
	defer buffer.Close()
	buffer.Concat(store.Prefix, ":", key)

	count, expiration := store.cache.Reset(buffer.String(), rate.Period)

	lctx := common.GetContextFromState(time.Now(), rate, expiration, count)
	return lctx, nil
}
