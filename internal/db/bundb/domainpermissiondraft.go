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
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

func (d *domainDB) getDomainPermissionDraft(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.DomainPermissionDraft) error,
	keyParts ...any,
) (*gtsmodel.DomainPermissionDraft, error) {
	// Fetch perm draft from database cache with loader callback.
	permDraft, err := d.state.Caches.DB.DomainPermissionDraft.LoadOne(
		lookup,
		// Only called if not cached.
		func() (*gtsmodel.DomainPermissionDraft, error) {
			var permDraft gtsmodel.DomainPermissionDraft
			if err := dbQuery(&permDraft); err != nil {
				return nil, err
			}
			return &permDraft, nil
		},
		keyParts...,
	)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// No need to fully populate.
		return permDraft, nil
	}

	if permDraft.CreatedByAccount == nil {
		// Not set, fetch from database.
		permDraft.CreatedByAccount, err = d.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			permDraft.CreatedByAccountID,
		)
		if err != nil {
			return nil, gtserror.Newf("error populating created by account: %w", err)
		}
	}

	return permDraft, nil
}

func (d *domainDB) GetDomainPermissionDraftByID(
	ctx context.Context,
	id string,
) (*gtsmodel.DomainPermissionDraft, error) {
	return d.getDomainPermissionDraft(
		ctx,
		"ID",
		func(permDraft *gtsmodel.DomainPermissionDraft) error {
			return d.db.
				NewSelect().
				Model(permDraft).
				Where("? = ?", bun.Ident("domain_permission_draft.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (d *domainDB) GetDomainPermissionDrafts(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
	permSubID string,
	domain string,
	page *paging.Page,
) (
	[]*gtsmodel.DomainPermissionDraft,
	error,
) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		permDraftIDs = make([]string, 0, limit)
	)

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_drafts"),
			bun.Ident("domain_permission_draft"),
		).
		// Select only IDs from table
		Column("domain_permission_draft.id")

	// Return only items with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where(
			"? < ?",
			bun.Ident("domain_permission_draft.id"),
			maxID,
		)
	}

	// Return only items with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where(
			"? > ?",
			bun.Ident("domain_permission_draft.id"),
			minID,
		)
	}

	// Return only items with
	// given permission type.
	if permType != gtsmodel.DomainPermissionUnknown {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.permission_type"),
			permType,
		)
	}

	// Return only items with
	// given subscription ID.
	if permSubID != "" {
		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.subscription_id"),
			permSubID,
		)
	}

	// Return only items
	// with given domain.
	if domain != "" {
		var err error

		// Normalize domain as punycode.
		domain, err = util.Punify(domain)
		if err != nil {
			return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
		}

		q = q.Where(
			"? = ?",
			bun.Ident("domain_permission_draft.domain"),
			domain,
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
			bun.Ident("domain_permission_draft.id"),
		)
	} else {
		// Page down.
		q = q.OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_draft.id"),
		)
	}

	if err := q.Scan(ctx, &permDraftIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(permDraftIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want items
	// to be sorted by ID desc, so reverse slice.
	if order == paging.OrderAscending {
		slices.Reverse(permDraftIDs)
	}

	// Allocate return slice (will be at most len permDraftIDs)
	permDrafts := make([]*gtsmodel.DomainPermissionDraft, 0, len(permDraftIDs))
	for _, id := range permDraftIDs {
		permDraft, err := d.GetDomainPermissionDraftByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission draft %q: %v", id, err)
			continue
		}

		// Append to return slice
		permDrafts = append(permDrafts, permDraft)
	}

	return permDrafts, nil
}

func (d *domainDB) PutDomainPermissionDraft(
	ctx context.Context,
	permDraft *gtsmodel.DomainPermissionDraft,
) error {
	var err error

	// Normalize the domain as punycode
	permDraft.Domain, err = util.Punify(permDraft.Domain)
	if err != nil {
		return gtserror.Newf("error punifying domain %s: %w", permDraft.Domain, err)
	}

	return d.state.Caches.DB.DomainPermissionDraft.Store(
		permDraft,
		func() error {
			_, err := d.db.
				NewInsert().
				Model(permDraft).
				Exec(ctx)
			return err
		},
	)
}

func (d *domainDB) DeleteDomainPermissionDraft(
	ctx context.Context,
	id string,
) error {
	// Delete the permDraft from DB.
	q := d.db.NewDelete().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_drafts"),
			bun.Ident("domain_permission_draft"),
		).
		Where(
			"? = ?",
			bun.Ident("domain_permission_draft.id"),
			id,
		)

	_, err := q.Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached model by ID.
	d.state.Caches.DB.DomainPermissionDraft.Invalidate("ID", id)

	return nil
}
