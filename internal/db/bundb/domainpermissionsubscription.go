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
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/uptrace/bun"
)

func (d *domainDB) getDomainPermissionSubscription(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.DomainPermissionSubscription) error,
	keyParts ...any,
) (*gtsmodel.DomainPermissionSubscription, error) {
	// Fetch perm subscription from database cache with loader callback.
	permSub, err := d.state.Caches.DB.DomainPermissionSubscription.LoadOne(
		lookup,
		// Only called if not cached.
		func() (*gtsmodel.DomainPermissionSubscription, error) {
			var permSub gtsmodel.DomainPermissionSubscription
			if err := dbQuery(&permSub); err != nil {
				return nil, err
			}
			return &permSub, nil
		},
		keyParts...,
	)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// No need to fully populate.
		return permSub, nil
	}

	if permSub.CreatedByAccount == nil {
		// Not set, fetch from database.
		permSub.CreatedByAccount, err = d.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			permSub.CreatedByAccountID,
		)
		if err != nil {
			return nil, gtserror.Newf("error populating created by account: %w", err)
		}
	}

	return permSub, nil
}

func (d *domainDB) GetDomainPermissionSubscriptionByID(
	ctx context.Context,
	id string,
) (*gtsmodel.DomainPermissionSubscription, error) {
	return d.getDomainPermissionSubscription(
		ctx,
		"ID",
		func(permSub *gtsmodel.DomainPermissionSubscription) error {
			return d.db.
				NewSelect().
				Model(permSub).
				Where("? = ?", bun.Ident("domain_permission_subscription.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (d *domainDB) GetDomainPermissionSubscriptions(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
	page *paging.Page,
) (
	[]*gtsmodel.DomainPermissionSubscription,
	error,
) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		permSubIDs = make([]string, 0, limit)
	)

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_subscriptions"),
			bun.Ident("domain_permission_subscription"),
		).
		// Select only IDs from table
		Column("domain_permission_subscription.id")

	// Return only items with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where(
			"? < ?",
			bun.Ident("domain_permission_subscription.id"),
			maxID,
		)
	}

	// Return only items with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where(
			"? > ?",
			bun.Ident("domain_permission_subscription.id"),
			minID,
		)
	}

	// Return only items with
	// given permission type.
	if permType != gtsmodel.DomainPermissionUnknown {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_subscription.permission_type"),
			permType,
		)
	}

	if limit > 0 {
		// Limit amount of
		// items returned.
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr(
			"? ASC",
			bun.Ident("domain_permission_subscription.id"),
		)
	} else {
		// Page down.
		q = q.OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_subscription.id"),
		)
	}

	if err := q.Scan(ctx, &permSubIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(permSubIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want items
	// to be sorted by ID desc, so reverse slice.
	if order == paging.OrderAscending {
		slices.Reverse(permSubIDs)
	}

	// Allocate return slice (will be at most len permSubIDs).
	permSubs := make([]*gtsmodel.DomainPermissionSubscription, 0, len(permSubIDs))
	for _, id := range permSubIDs {
		permSub, err := d.GetDomainPermissionSubscriptionByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission subscription %q: %v", id, err)
			continue
		}

		// Append to return slice
		permSubs = append(permSubs, permSub)
	}

	return permSubs, nil
}

func (d *domainDB) GetDomainPermissionSubscriptionsByPriority(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
) (
	[]*gtsmodel.DomainPermissionSubscription,
	error,
) {
	permSubIDs := []string{}

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_subscriptions"),
			bun.Ident("domain_permission_subscription"),
		).
		// Select only IDs from table
		Column("domain_permission_subscription.id").
		// Select only subs of given perm type.
		Where(
			"? = ?",
			bun.Ident("domain_permission_subscription.permission_type"),
			permType,
		).
		// Order by priority descending.
		OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_subscription.priority"),
		)

	if err := q.Scan(ctx, &permSubIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(permSubIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// Allocate return slice (will be at most len permSubIDs).
	permSubs := make([]*gtsmodel.DomainPermissionSubscription, 0, len(permSubIDs))
	for _, id := range permSubIDs {
		permSub, err := d.GetDomainPermissionSubscriptionByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission subscription %q: %v", id, err)
			continue
		}

		// Append to return slice
		permSubs = append(permSubs, permSub)
	}

	return permSubs, nil
}

func (d *domainDB) PutDomainPermissionSubscription(
	ctx context.Context,
	permSubscription *gtsmodel.DomainPermissionSubscription,
) error {
	return d.state.Caches.DB.DomainPermissionSubscription.Store(
		permSubscription,
		func() error {
			_, err := d.db.
				NewInsert().
				Model(permSubscription).
				Exec(ctx)
			return err
		},
	)
}

func (d *domainDB) UpdateDomainPermissionSubscription(
	ctx context.Context,
	permSubscription *gtsmodel.DomainPermissionSubscription,
	columns ...string,
) error {
	return d.state.Caches.DB.DomainPermissionSubscription.Store(
		permSubscription,
		func() error {
			_, err := d.db.
				NewUpdate().
				Model(permSubscription).
				Where("? = ?", bun.Ident("id"), permSubscription.ID).
				Column(columns...).
				Exec(ctx)
			return err
		},
	)
}

func (d *domainDB) DeleteDomainPermissionSubscription(
	ctx context.Context,
	id string,
) error {
	// Delete the permSub from DB.
	q := d.db.NewDelete().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_subscriptions"),
			bun.Ident("domain_permission_subscription"),
		).
		Where(
			"? = ?",
			bun.Ident("domain_permission_subscription.id"),
			id,
		)

	_, err := q.Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached model by ID.
	d.state.Caches.DB.DomainPermissionSubscription.Invalidate("ID", id)

	return nil
}

func (d *domainDB) CountDomainPermissionSubscriptionPerms(
	ctx context.Context,
	id string,
) (int, error) {
	permSubscription, err := d.GetDomainPermissionSubscriptionByID(
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil {
		return 0, err
	}

	q := d.db.NewSelect()

	if permSubscription.PermissionType == gtsmodel.DomainPermissionBlock {
		q = q.TableExpr(
			"? AS ?",
			bun.Ident("domain_blocks"),
			bun.Ident("perm"),
		)
	} else {
		q = q.TableExpr(
			"? AS ?",
			bun.Ident("domain_allows"),
			bun.Ident("perm"),
		)
	}

	return q.
		Column("perm.id").
		Where("? = ?", bun.Ident("perm.subscription_id"), id).
		Count(ctx)
}
