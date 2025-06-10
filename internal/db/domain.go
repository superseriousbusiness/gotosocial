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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// Domain contains DB functions related to domains and domain blocks.
type Domain interface {
	/*
		Block/allow storage + retrieval functions.
	*/

	// PutDomainAllow puts the given instance-level domain allow into the database.
	PutDomainAllow(ctx context.Context, allow *gtsmodel.DomainAllow) error

	// GetDomainAllow returns one instance-level domain allow with the given domain, if it exists.
	GetDomainAllow(ctx context.Context, domain string) (*gtsmodel.DomainAllow, error)

	// GetDomainAllowByID returns one instance-level domain allow with the given id, if it exists.
	GetDomainAllowByID(ctx context.Context, id string) (*gtsmodel.DomainAllow, error)

	// GetDomainAllows returns all instance-level domain allows currently enforced by this instance.
	GetDomainAllows(ctx context.Context) ([]*gtsmodel.DomainAllow, error)

	// GetDomainAllowsBySubscriptionID gets all domain allows that have the given subscription ID.
	GetDomainAllowsBySubscriptionID(ctx context.Context, subscriptionID string) ([]*gtsmodel.DomainAllow, error)

	// UpdateDomainAllow updates the given domain allow, setting the provided columns (empty for all).
	UpdateDomainAllow(ctx context.Context, allow *gtsmodel.DomainAllow, columns ...string) error

	// DeleteDomainAllow deletes an instance-level domain allow with the given domain, if it exists.
	DeleteDomainAllow(ctx context.Context, domain string) error

	// PutDomainBlock puts the given instance-level domain block into the database.
	PutDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) error

	// GetDomainBlock returns one instance-level domain block with the given domain, if it exists.
	GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, error)

	// GetDomainBlockByID returns one instance-level domain block with the given id, if it exists.
	GetDomainBlockByID(ctx context.Context, id string) (*gtsmodel.DomainBlock, error)

	// GetDomainBlocksBySubscriptionID gets all domain blocks that have the given subscription ID.
	GetDomainBlocksBySubscriptionID(ctx context.Context, subscriptionID string) ([]*gtsmodel.DomainBlock, error)

	// GetDomainBlocks returns all instance-level domain blocks currently enforced by this instance.
	GetDomainBlocks(ctx context.Context) ([]*gtsmodel.DomainBlock, error)

	// UpdateDomainBlock updates the given domain block, setting the provided columns (empty for all).
	UpdateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock, columns ...string) error

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

	/*
		Domain permission draft stuff.
	*/

	// GetDomainPermissionDraftByID gets one DomainPermissionDraft with the given ID.
	GetDomainPermissionDraftByID(ctx context.Context, id string) (*gtsmodel.DomainPermissionDraft, error)

	// GetDomainPermissionDrafts returns a page of
	// DomainPermissionDrafts using the given parameters.
	GetDomainPermissionDrafts(
		ctx context.Context,
		permType gtsmodel.DomainPermissionType,
		permSubID string,
		domain string,
		page *paging.Page,
	) ([]*gtsmodel.DomainPermissionDraft, error)

	// PutDomainPermissionDraft stores one DomainPermissionDraft.
	PutDomainPermissionDraft(ctx context.Context, permDraft *gtsmodel.DomainPermissionDraft) error

	// DeleteDomainPermissionDraft deletes one DomainPermissionDraft with the given id.
	DeleteDomainPermissionDraft(ctx context.Context, id string) error

	/*
		Domain permission exclude stuff.
	*/

	// GetDomainPermissionExcludeByID gets one DomainPermissionExclude with the given ID.
	GetDomainPermissionExcludeByID(ctx context.Context, id string) (*gtsmodel.DomainPermissionExclude, error)

	// GetDomainPermissionExcludes returns a page of
	// DomainPermissionExcludes using the given parameters.
	GetDomainPermissionExcludes(
		ctx context.Context,
		domain string,
		page *paging.Page,
	) ([]*gtsmodel.DomainPermissionExclude, error)

	// PutDomainPermissionExclude stores one DomainPermissionExclude.
	PutDomainPermissionExclude(ctx context.Context, permExclude *gtsmodel.DomainPermissionExclude) error

	// DeleteDomainPermissionExclude deletes one DomainPermissionExclude with the given id.
	DeleteDomainPermissionExclude(ctx context.Context, id string) error

	// IsDomainPermissionExcluded returns true if the given domain matches in the list of excluded domains.
	IsDomainPermissionExcluded(ctx context.Context, domain string) (bool, error)

	/*
		Domain permission subscription stuff.
	*/

	// GetDomainPermissionSubscriptionByID gets one DomainPermissionSubscription with the given ID.
	GetDomainPermissionSubscriptionByID(ctx context.Context, id string) (*gtsmodel.DomainPermissionSubscription, error)

	// GetDomainPermissionSubscriptions returns a page of
	// DomainPermissionSubscriptions using the given parameters.
	GetDomainPermissionSubscriptions(
		ctx context.Context,
		permType gtsmodel.DomainPermissionType,
		page *paging.Page,
	) ([]*gtsmodel.DomainPermissionSubscription, error)

	// GetDomainPermissionSubscriptionsByPriority returns *all* domain permission
	// subscriptions of the given permission type, sorted by priority descending.
	GetDomainPermissionSubscriptionsByPriority(
		ctx context.Context,
		permType gtsmodel.DomainPermissionType,
	) ([]*gtsmodel.DomainPermissionSubscription, error)

	// PutDomainPermissionSubscription stores one DomainPermissionSubscription.
	PutDomainPermissionSubscription(ctx context.Context, permSub *gtsmodel.DomainPermissionSubscription) error

	// UpdateDomainPermissionSubscription updates the provided
	// columns of one DomainPermissionSubscription.
	UpdateDomainPermissionSubscription(
		ctx context.Context,
		permSub *gtsmodel.DomainPermissionSubscription,
		columns ...string,
	) error

	// DeleteDomainPermissionSubscription deletes one DomainPermissionSubscription with the given id.
	DeleteDomainPermissionSubscription(ctx context.Context, id string) error

	// CountDomainPermissionSubscriptionPerms counts the number of permissions
	// currently managed by the domain permission subscription of the given ID.
	CountDomainPermissionSubscriptionPerms(ctx context.Context, id string) (int, error)
}
