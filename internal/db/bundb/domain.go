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

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"golang.org/x/net/idna"
)

type domainDB struct {
	conn  *DBConn
	state *state.State
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

	// Normalize the domain as punycode
	block.Domain, err = normalizeDomain(block.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain in DB
	if _, err := d.conn.NewInsert().
		Model(block).
		Exec(ctx); err != nil {
		return d.conn.ProcessError(err)
	}

	// Clear the domain block cache (for later reload)
	d.state.Caches.GTS.DomainBlock().Clear()

	return nil
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, db.Error) {
	var err error

	// Normalize the domain as punycode
	domain, err = normalizeDomain(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return nil, db.ErrNoEntries
	}

	var block gtsmodel.DomainBlock

	// Look for block matching domain in DB
	q := d.conn.
		NewSelect().
		Model(&block).
		Where("? = ?", bun.Ident("domain_block.domain"), domain)
	if err := q.Scan(ctx); err != nil {
		return nil, d.conn.ProcessError(err)
	}

	return &block, nil
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

	// Clear the domain block cache (for later reload)
	d.state.Caches.GTS.DomainBlock().Clear()

	return nil
}

func (d *domainDB) IsDomainBlocked(ctx context.Context, domain string) (bool, db.Error) {
	// Normalize the domain as punycode
	domain, err := normalizeDomain(domain)
	if err != nil {
		return false, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return false, nil
	}

	// Check the cache for a domain block (hydrating the cache with callback if necessary)
	return d.state.Caches.GTS.DomainBlock().IsBlocked(domain, func() ([]string, error) {
		var domains []string

		// Scan list of all blocked domains from DB
		q := d.conn.NewSelect().
			Table("domain_blocks").
			Column("domain")
		if err := q.Scan(ctx, &domains); err != nil {
			return nil, d.conn.ProcessError(err)
		}

		return domains, nil
	})
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
