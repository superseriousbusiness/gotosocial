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
	/*
		Block/allow storage + retrieval functions.
	*/

	// CreateDomainAllow puts the given instance-level domain allow into the database.
	CreateDomainAllow(ctx context.Context, allow *gtsmodel.DomainAllow) error

	// GetDomainAllow returns one instance-level domain allow with the given domain, if it exists.
	GetDomainAllow(ctx context.Context, domain string) (*gtsmodel.DomainAllow, error)

	// GetDomainAllowByID returns one instance-level domain allow with the given id, if it exists.
	GetDomainAllowByID(ctx context.Context, id string) (*gtsmodel.DomainAllow, error)

	// GetDomainAllows returns all instance-level domain allows currently enforced by this instance.
	GetDomainAllows(ctx context.Context) ([]*gtsmodel.DomainAllow, error)

	// DeleteDomainAllow deletes an instance-level domain allow with the given domain, if it exists.
	DeleteDomainAllow(ctx context.Context, domain string) error

	// CreateDomainBlock puts the given instance-level domain block into the database.
	CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) error

	// GetDomainBlock returns one instance-level domain block with the given domain, if it exists.
	GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, error)

	// GetDomainBlockByID returns one instance-level domain block with the given id, if it exists.
	GetDomainBlockByID(ctx context.Context, id string) (*gtsmodel.DomainBlock, error)

	// GetDomainBlocks returns all instance-level domain blocks currently enforced by this instance.
	GetDomainBlocks(ctx context.Context) ([]*gtsmodel.DomainBlock, error)

	// DeleteDomainBlock deletes an instance-level domain block with the given domain, if it exists.
	DeleteDomainBlock(ctx context.Context, domain string) error

	/*
		Block/allow checking functions.
	*/

	// IsDomainBlocked checks if domain is blocked, accounting for both explicit allows and blocks.
	// Will check allows first, so an allowed domain will always return false, even if it's also blocked.
	IsDomainBlocked(ctx context.Context, domain string) (bool, error)

	// AreDomainsBlocked calls IsDomainBlocked for each domain.
	// Will return true if even one of the given domains is blocked.
	AreDomainsBlocked(ctx context.Context, domains []string) (bool, error)

	// IsURIBlocked calls IsDomainBlocked for the host of the given URI.
	IsURIBlocked(ctx context.Context, uri *url.URL) (bool, error)

	// AreURIsBlocked calls IsURIBlocked for each URI.
	// Will return true if even one of the given URIs is blocked.
	AreURIsBlocked(ctx context.Context, uris []*url.URL) (bool, error)
}
