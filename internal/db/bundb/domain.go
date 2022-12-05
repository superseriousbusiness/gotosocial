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
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/miekg/dns"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"golang.org/x/net/idna"
)

// maxDomainParts is the maximum number of subdomain parts in
// domain that we accept both going into the database, and during
// fetch attempts to reduce risk of denial of service.
const maxDomainParts = 5

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

	if dns.CountLabel(block.Domain) > maxDomainParts {
		return errors.New("invalid domain: contains too many subdomain parts")
	}

	return d.state.Caches.GTS.DomainBlock().Store(block, func() error {
		_, err := d.conn.NewInsert().
			Model(block).
			Exec(ctx)
		return d.conn.ProcessError(err)
	})
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, db.Error) {
	var err error

	// Normalize the domain as punycode
	domain, err = normalizeDomain(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() {
		return nil, db.ErrNoEntries
	}

	// Split domain into constituent parts
	parts := dns.SplitDomainName(domain)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid domain: %s", domain)
	}

	// Create first domain lookup attempt (this is root level + TLD)
	lookup := parts[len(parts)-2] + "." + parts[len(parts)-1]
	parts = parts[:len(parts)-2]

	for i := 0; i < maxDomainParts; i++ {
		// Check if this lookup result already cached
		cached := d.state.Caches.GTS.DomainBlock().Has("Domain", lookup)

		// Attempt to fetch domain block for lookup
		block, err := d.getDomainBlock(ctx, lookup)
		if err != nil {
			// We return early if this is one of:
			// - NOT db.ErrNoEntries, i.e. unrecoverable error
			// - cached db.ErrNoEntries, i.e. have tried this before
			if cached || !errors.Is(err, db.ErrNoEntries) {
				return nil, err
			}
		}

		if block != nil {
			// A block was found!
			return block, nil
		}

		if len(parts) == 0 {
			// No further domain parts to append
			return nil, db.ErrNoEntries
		}

		// Prepend the next subdomain part to lookup
		lookup = parts[len(parts)-1] + "." + lookup
		parts = parts[:len(parts)-1]
	}

	return nil, db.ErrNoEntries
}

func (d *domainDB) getDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, error) {
	return d.state.Caches.GTS.DomainBlock().Load("Domain", func() (*gtsmodel.DomainBlock, error) {
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
	d.state.Caches.GTS.DomainBlock().Invalidate(domain)

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
