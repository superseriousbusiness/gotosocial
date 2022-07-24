/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"context"
	"database/sql"
	"net/url"
	"strings"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type domainDB struct {
	conn  *DBConn
	cache cache.Cache[string, *gtsmodel.DomainBlock]
}

func (d *domainDB) CreateDomainBlock(ctx context.Context, block gtsmodel.DomainBlock) db.Error {
	// Normalize to lowercase
	block.Domain = strings.ToLower(block.Domain)

	// Attempt to insert new domain block
	_, err := d.conn.NewInsert().
		Model(&block).
		Exec(ctx, &block)
	if err != nil {
		return d.conn.ProcessError(err)
	}

	// Cache this domain block
	d.cache.Set(block.Domain, &block)

	return nil
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, db.Error) {
	// Normalize to lowercase
	domain = strings.ToLower(domain)

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() {
		return nil, db.ErrNoEntries
	}

	// Check for already cached rblock
	if block, ok := d.cache.Get(domain); ok {
		// A 'nil' return value is a sentinel value for no block
		if block == nil {
			return nil, db.ErrNoEntries
		}

		// Else, this block exists
		return block, nil
	}

	block := &gtsmodel.DomainBlock{}

	q := d.conn.
		NewSelect().
		Model(block).
		Where("domain = ?", domain).
		Limit(1)

	// Query database for domain block
	switch err := q.Scan(ctx); err {
	// No error, block found
	case nil:
		d.cache.Set(domain, block)
		return block, nil

	// No error, simply not found
	case sql.ErrNoRows:
		d.cache.Set(domain, nil)
		return nil, db.ErrNoEntries

	// Any other db error
	default:
		return nil, d.conn.ProcessError(err)
	}
}

func (d *domainDB) DeleteDomainBlock(ctx context.Context, domain string) db.Error {
	// Normalize to lowercase
	domain = strings.ToLower(domain)

	// Attempt to delete domain block
	_, err := d.conn.NewDelete().
		Model((*gtsmodel.DomainBlock)(nil)).
		Where("domain = ?", domain).
		Exec(ctx, nil)
	if err != nil {
		return d.conn.ProcessError(err)
	}

	// Clear domain from cache
	d.cache.Invalidate(domain)

	return nil
}

func (d *domainDB) IsDomainBlocked(ctx context.Context, domain string) (bool, db.Error) {
	block, err := d.GetDomainBlock(ctx, domain)
	if err == nil || err == db.ErrNoEntries {
		return (block != nil), nil
	}
	return false, err
}

func (d *domainDB) AreDomainsBlocked(ctx context.Context, domains []string) (bool, db.Error) {
	for _, domain := range domains {
		if blocked, err := d.IsDomainBlocked(ctx, domain); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}

func (d *domainDB) IsURIBlocked(ctx context.Context, uri *url.URL) (bool, db.Error) {
	return d.IsDomainBlocked(ctx, uri.Hostname())
}

func (d *domainDB) AreURIsBlocked(ctx context.Context, uris []*url.URL) (bool, db.Error) {
	for _, uri := range uris {
		if blocked, err := d.IsDomainBlocked(ctx, uri.Hostname()); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}
