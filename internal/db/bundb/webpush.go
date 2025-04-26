// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bundb

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	webpushgo "github.com/SherClockHolmes/webpush-go"
	"github.com/uptrace/bun"
)

type webPushDB struct {
	db    *bun.DB
	state *state.State
}

func (w *webPushDB) GetVAPIDKeyPair(ctx context.Context) (*gtsmodel.VAPIDKeyPair, error) {
	var err error

	vapidKeyPair, err := w.getVAPIDKeyPair(ctx)
	if err != nil {
		return nil, err
	}
	if vapidKeyPair != nil {
		return vapidKeyPair, nil
	}

	// If there aren't any, generate new ones.
	vapidKeyPair = &gtsmodel.VAPIDKeyPair{}
	if vapidKeyPair.Private, vapidKeyPair.Public, err = webpushgo.GenerateVAPIDKeys(); err != nil {
		return nil, gtserror.Newf("error generating VAPID key pair: %w", err)
	}

	// Store the keys in the database.
	if _, err = w.db.NewInsert().
		Model(vapidKeyPair).
		Exec(ctx); // nocollapse
	err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// Multiple concurrent attempts to generate new keys, and this one didn't win.
			// Get the results of the one that did.
			return w.getVAPIDKeyPair(ctx)
		}
		return nil, err
	}

	// Cache the keys.
	w.state.Caches.DB.VAPIDKeyPair.Store(vapidKeyPair)

	return vapidKeyPair, nil
}

// getVAPIDKeyPair gets an existing VAPID key pair from cache or DB.
// If there is no existing VAPID key pair, it returns nil, with no error.
func (w *webPushDB) getVAPIDKeyPair(ctx context.Context) (*gtsmodel.VAPIDKeyPair, error) {
	// Look for cached keys.
	vapidKeyPair := w.state.Caches.DB.VAPIDKeyPair.Load()
	if vapidKeyPair != nil {
		return vapidKeyPair, nil
	}

	// Look for previously generated keys in the database.
	vapidKeyPair = &gtsmodel.VAPIDKeyPair{}
	if err := w.db.NewSelect().
		Model(vapidKeyPair).
		Limit(1).
		Scan(ctx); // nocollapse
	err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, err
	}

	return vapidKeyPair, nil
}

func (w *webPushDB) DeleteVAPIDKeyPair(ctx context.Context) error {
	// Delete any existing keys.
	if _, err := w.db.NewTruncateTable().
		Model((*gtsmodel.VAPIDKeyPair)(nil)).
		Exec(ctx); // nocollapse
	err != nil {
		return err
	}

	// Clear the key cache.
	w.state.Caches.DB.VAPIDKeyPair.Store(nil)

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
	if err != nil {
		return nil, err
	}
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
