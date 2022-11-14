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
	"net/url"
	"strings"
	"time"

	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
	"golang.org/x/net/idna"
)

type domainDB struct {
	conn  *DBConn
	cache *result.Cache[*gtsmodel.DomainBlock]
}

func (d *domainDB) init() {
	// Initialize domain block result cache
	d.cache = result.NewSized([]result.Lookup{
		{Name: "Domain"},
	}, func(d1 *gtsmodel.DomainBlock) *gtsmodel.DomainBlock {
		d2 := new(gtsmodel.DomainBlock)
		*d2 = *d1
		return d2
	}, 1000)

	// Set cache TTL and start sweep routine
	d.cache.SetTTL(time.Minute*5, false)
	d.cache.Start(time.Second * 10)
}

// normalizeDomain converts the given domain to lowercase
// then to punycode (for international domain names).
//
// Returns the resulting domain or an error if the
// punycode conversion fails.
func normalizeDomain(domain string) (out string, err error) {
	out = strings.ToLower(domain)
	out, err = idna.ToASCII(out)
	return out, err
}

func (d *domainDB) CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) db.Error {
	var err error

	block.Domain, err = normalizeDomain(block.Domain)
	if err != nil {
		return err
	}

	return d.cache.Store(block, func() error {
		_, err := d.conn.NewInsert().
			Model(block).
			Exec(ctx)
		return d.conn.ProcessError(err)
	})
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, db.Error) {
	var err error

	domain, err = normalizeDomain(domain)
	if err != nil {
		return nil, err
	}

	return d.cache.Load("Domain", func() (*gtsmodel.DomainBlock, error) {
		// Check for easy case, domain referencing *us*
		if domain == "" || domain == config.GetAccountDomain() {
			return nil, db.ErrNoEntries
		}

		var block gtsmodel.DomainBlock

		q := d.conn.
			NewSelect().
			Model(&block).
			Where("? = ?", bun.Ident("domain_block.domain"), domain).
			Limit(1)
		if err := q.Scan(ctx); err != nil {
			return nil, d.conn.ProcessError(err)
		}

		return &block, nil
	}, domain)
}

func (d *domainDB) DeleteDomainBlock(ctx context.Context, domain string) db.Error {
	var err error

	domain, err = normalizeDomain(domain)
	if err != nil {
		return err
	}

	// Attempt to delete domain block
	if _, err := d.conn.NewDelete().
		Model((*gtsmodel.DomainBlock)(nil)).
		Where("? = ?", bun.Ident("domain_block.domain"), domain).
		Exec(ctx); err != nil {
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
