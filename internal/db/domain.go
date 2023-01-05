/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package db

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Domain contains DB functions related to domains and domain blocks.
type Domain interface {
	// CreateDomainBlock ...
	CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) Error

	// GetDomainBlock ...
	GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, Error)

	// DeleteDomainBlock ...
	DeleteDomainBlock(ctx context.Context, domain string) Error

	// IsDomainBlocked checks if an instance-level domain block exists for the given domain string (eg., `example.org`).
	IsDomainBlocked(ctx context.Context, domain string) (bool, Error)

	// AreDomainsBlocked checks if an instance-level domain block exists for any of the given domains strings, and returns true if even one is found.
	AreDomainsBlocked(ctx context.Context, domains []string) (bool, Error)

	// IsURIBlocked checks if an instance-level domain block exists for the `host` in the given URI (eg., `https://example.org/users/whatever`).
	IsURIBlocked(ctx context.Context, uri *url.URL) (bool, Error)

	// AreURIsBlocked checks if an instance-level domain block exists for any `host` in the given URI slice, and returns true if even one is found.
	AreURIsBlocked(ctx context.Context, uris []*url.URL) (bool, Error)
}
