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

package db

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Domain contains DB functions related to domains and domain blocks.
type Domain interface {
	// CreateDomainBlock puts the given instance-level domain block into the database.
	CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) Error

	// GetDomainBlock returns one instance-level domain block with the given domain, if it exists.
	GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, Error)

	// GetDomainBlockByID returns one instance-level domain block with the given id, if it exists.
	GetDomainBlockByID(ctx context.Context, id string) (*gtsmodel.DomainBlock, Error)

	// GetDomainBlocks returns all instance-level domain blocks currently enforced by this instance.
	GetDomainBlocks(ctx context.Context) ([]*gtsmodel.DomainBlock, error)

	// DeleteDomainBlock deletes an instance-level domain block with the given domain, if it exists.
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
