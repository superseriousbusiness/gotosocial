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

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

func (d *domainDB) PutDomainPermissionExclude(
	ctx context.Context,
	exclude *gtsmodel.DomainPermissionExclude,
) error {
	// Normalize the domain as punycode
	var err error
	exclude.Domain, err = util.Punify(exclude.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain perm exclude in DB
	if _, err := d.db.NewInsert().
		Model(exclude).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain perm exclude cache (for later reload)
	d.state.Caches.DB.DomainPermissionExclude.Clear()

	return nil
}

func (d *domainDB) IsDomainPermissionExcluded(ctx context.Context, domain string) (bool, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return false, err
	}

	// Check if our host and given domain are equal
	// or part of the same second-level domain; we
	// always exclude such perms as creating blocks
	// or allows in such cases may break things.
	if dns.CompareDomainName(domain, config.GetHost()) >= 2 {
		return true, nil
	}

	// Func to scan list of all
	// excluded domain perms from DB.
	loadF := func() ([]string, error) {
		var domains []string

		if err := d.db.
			NewSelect().
			Table("domain_excludes").
			Column("domain").
			Scan(ctx, &domains); err != nil {
			return nil, err
		}

		return domains, nil
	}

	// Check the cache for a domain perm exclude,
	// hydrating the cache with loadF if necessary.
	return d.state.Caches.DB.DomainPermissionExclude.Matches(domain, loadF)
}

func (d *domainDB) GetDomainPermissionExcludeByID(
	ctx context.Context,
	id string,
) (*gtsmodel.DomainPermissionExclude, error) {
	exclude := new(gtsmodel.DomainPermissionExclude)

	q := d.db.
		NewSelect().
		Model(exclude).
		Where("? = ?", bun.Ident("domain_permission_exclude.id"), id)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// No need to fully populate.
		return exclude, nil
	}

	if exclude.CreatedByAccount == nil {
		// Not set, fetch from database.
		var err error
		exclude.CreatedByAccount, err = d.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			exclude.CreatedByAccountID,
		)
		if err != nil {
			return nil, gtserror.Newf("error populating created by account: %w", err)
		}
	}

	return exclude, nil
}

func (d *domainDB) GetDomainPermissionExcludes(
	ctx context.Context,
	domain string,
	page *paging.Page,
) (
	[]*gtsmodel.DomainPermissionExclude,
	error,
) {
	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		excludeIDs = make([]string, 0, limit)
	)

	q := d.db.
		NewSelect().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_excludes"),
			bun.Ident("domain_permission_exclude"),
		).
		// Select only IDs from table
		Column("domain_permission_exclude.id")

	// Return only items with id
	// lower than provided maxID.
	if maxID != "" {
		q = q.Where(
			"? < ?",
			bun.Ident("domain_permission_exclude.id"),
			maxID,
		)
	}

	// Return only items with id
	// greater than provided minID.
	if minID != "" {
		q = q.Where(
			"? > ?",
			bun.Ident("domain_permission_exclude.id"),
			minID,
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
			bun.Ident("domain_permission_exclude.domain"),
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
			bun.Ident("domain_permission_exclude.id"),
		)
	} else {
		// Page down.
		q = q.OrderExpr(
			"? DESC",
			bun.Ident("domain_permission_exclude.id"),
		)
	}

	if err := q.Scan(ctx, &excludeIDs); err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(excludeIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want items
	// to be sorted by ID desc, so reverse slice.
	if order == paging.OrderAscending {
		slices.Reverse(excludeIDs)
	}

	// Allocate return slice (will be at most len permSubIDs).
	excludes := make([]*gtsmodel.DomainPermissionExclude, 0, len(excludeIDs))
	for _, id := range excludeIDs {
		exclude, err := d.GetDomainPermissionExcludeByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting domain permission exclude %q: %v", id, err)
			continue
		}

		// Append to return slice
		excludes = append(excludes, exclude)
	}

	return excludes, nil
}

func (d *domainDB) DeleteDomainPermissionExclude(
	ctx context.Context,
	id string,
) error {
	// Delete the permSub from DB.
	q := d.db.NewDelete().
		TableExpr(
			"? AS ?",
			bun.Ident("domain_permission_excludes"),
			bun.Ident("domain_permission_exclude"),
		).
		Where(
			"? = ?",
			bun.Ident("domain_permission_exclude.id"),
			id,
		)

	_, err := q.Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Clear the domain perm exclude cache (for later reload)
	d.state.Caches.DB.DomainPermissionExclude.Clear()

	return nil
}
