package bundb

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type webPushDB struct {
	db    *bun.DB
	state *state.State
}

func (w *webPushDB) GetVAPIDKeyPair(ctx context.Context) (*gtsmodel.VAPIDKeyPair, error) {
	// Look for cached keys.
	vapidKeyPair := w.state.Caches.DB.VAPIDKeyPair.Load()
	if vapidKeyPair != nil {
		return vapidKeyPair, nil
	}

	// Look for previously generated keys in the database.
	if err := w.db.NewSelect().
		Model(vapidKeyPair).
		Limit(1).
		Scan(ctx); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, err
	}

	// Cache the keys.
	w.state.Caches.DB.VAPIDKeyPair.Store(vapidKeyPair)

	return vapidKeyPair, nil
}

func (w *webPushDB) PutVAPIDKeyPair(ctx context.Context, vapidKeyPair *gtsmodel.VAPIDKeyPair) error {
	// Store the keys in the database.
	if _, err := w.db.NewInsert().
		Model(vapidKeyPair).
		Exec(ctx); // nocollapse
	err != nil {
		return err
	}

	// Cache the keys.
	w.state.Caches.DB.VAPIDKeyPair.Store(vapidKeyPair)

	return nil
}

func (w *webPushDB) GetWebPushSubscriptionByTokenID(ctx context.Context, tokenID string) (*gtsmodel.WebPushSubscription, error) {
	subscription, err := w.state.Caches.DB.WebPushSubscription.LoadOne(
		"TokenID",
		func() (*gtsmodel.WebPushSubscription, error) {
			var subscription gtsmodel.WebPushSubscription
			err := w.db.
				NewSelect().
				Model(&subscription).
				Where("? = ?", bun.Ident("token_id"), tokenID).
				Scan(ctx)
			return &subscription, err
		},
		tokenID,
	)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (w *webPushDB) PutWebPushSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription) error {
	return w.state.Caches.DB.WebPushSubscription.Store(subscription, func() error {
		_, err := w.db.NewInsert().
			Model(subscription).
			Exec(ctx)
		return err
	})
}

func (w *webPushDB) UpdateWebPushSubscription(ctx context.Context, subscription *gtsmodel.WebPushSubscription, columns ...string) error {
	// If we're updating by column, ensure "updated_at" is included.
	if len(columns) > 0 {
		columns = append(columns, "updated_at")
	}

	// Update database.
	result, err := w.db.
		NewUpdate().
		Model(subscription).
		Column(columns...).
		Where("? = ?", bun.Ident("id"), subscription.ID).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return gtserror.Newf("error getting updated row count: %w", err)
	}
	if rowsAffected == 0 {
		return db.ErrNoEntries
	}

	// Update cache.
	w.state.Caches.DB.WebPushSubscription.Put(subscription)

	return nil
}

func (w *webPushDB) DeleteWebPushSubscriptionByTokenID(ctx context.Context, tokenID string) error {
	// Deleted partial model for cache invalidation.
	var deleted gtsmodel.WebPushSubscription

	// Delete subscription, returning subset of columns used by invalidation hook.
	if _, err := w.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("token_id"), tokenID).
		Returning("?", bun.Ident("account_id")).
		Exec(ctx); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached subscription by token ID.
	w.state.Caches.DB.WebPushSubscription.Invalidate("TokenID", tokenID)

	// Call invalidate hook directly.
	w.state.Caches.OnInvalidateWebPushSubscription(&deleted)

	return nil
}

func (w *webPushDB) GetWebPushSubscriptionsByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.WebPushSubscription, error) {
	// Fetch IDs of all subscriptions created by this account.
	subscriptionIDs, err := loadPagedIDs(&w.state.Caches.DB.WebPushSubscriptionIDs, accountID, nil, func() ([]string, error) {
		// Subscription IDs not in cache. Perform DB query.
		var subscriptionIDs []string
		if _, err := w.db.
			NewSelect().
			Model((*gtsmodel.WebPushSubscription)(nil)).
			Column("id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			Order("id DESC").
			Exec(ctx, &subscriptionIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}
		return subscriptionIDs, nil
	})
	if len(subscriptionIDs) == 0 {
		return nil, nil
	}

	// Get each subscription by ID from the cache or DB.
	subscriptions, err := w.state.Caches.DB.WebPushSubscription.LoadIDs("ID",
		subscriptionIDs,
		func(uncached []string) ([]*gtsmodel.WebPushSubscription, error) {
			subscriptions := make([]*gtsmodel.WebPushSubscription, 0, len(uncached))
			if err := w.db.
				NewSelect().
				Model(&subscriptions).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); // nocollapse
			err != nil {
				return nil, err
			}
			return subscriptions, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Put the subscription structs in the same order as the filter IDs.
	xslices.OrderBy(
		subscriptions,
		subscriptionIDs,
		func(subscription *gtsmodel.WebPushSubscription) string {
			return subscription.ID
		},
	)

	return subscriptions, nil
}

func (w *webPushDB) DeleteWebPushSubscriptionsByAccountID(ctx context.Context, accountID string) error {
	// Deleted partial models for cache invalidation.
	var deleted []*gtsmodel.WebPushSubscription

	// Delete subscriptions, returning subset of columns.
	if _, err := w.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Returning("?", bun.Ident("account_id")).
		Exec(ctx); // nocollapse
	err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached subscriptions by account ID.
	w.state.Caches.DB.WebPushSubscription.Invalidate("AccountID", accountID)

	// Call invalidate hooks directly in case those entries weren't cached.
	for _, subscription := range deleted {
		w.state.Caches.OnInvalidateWebPushSubscription(subscription)
	}

	return nil
}
