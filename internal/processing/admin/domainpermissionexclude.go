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

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

func (p *Processor) DomainPermissionExcludeCreate(
	ctx context.Context,
	acct *gtsmodel.Account,
	domain string,
	privateComment string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permExclude := &gtsmodel.DomainPermissionExclude{
		ID:                 id.NewULID(),
		Domain:             domain,
		CreatedByAccountID: acct.ID,
		CreatedByAccount:   acct,
		PrivateComment:     privateComment,
	}

	if err := p.state.DB.PutDomainPermissionExclude(ctx, permExclude); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			const text = "a domain permission exclude already exists with this permission type and domain"
			err := fmt.Errorf("%w: %s", err, text)
			return nil, gtserror.NewErrorConflict(err, text)
		}

		// Real error.
		err := gtserror.Newf("db error putting domain permission exclude: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPerm(ctx, permExclude, false)
}

// DomainPermissionExcludeGet returns one
// domain permission exclude with the given id.
func (p *Processor) DomainPermissionExcludeGet(
	ctx context.Context,
	id string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permExclude, err := p.state.DB.GetDomainPermissionExcludeByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission exclude %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permExclude == nil {
		err := fmt.Errorf("domain permission exclude %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	return p.apiDomainPerm(ctx, permExclude, false)
}

// DomainPermissionExcludesGet returns a page of
// DomainPermissionExcludes with the given parameters.
func (p *Processor) DomainPermissionExcludesGet(
	ctx context.Context,
	domain string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	permExcludes, err := p.state.DB.GetDomainPermissionExcludes(
		ctx,
		domain,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(permExcludes)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := permExcludes[count-1].ID
	hi := permExcludes[0].ID

	// Convert each perm exclude to API model.
	items := make([]any, len(permExcludes))
	for i, permExclude := range permExcludes {
		apiPermExclude, err := p.apiDomainPerm(ctx, permExclude, false)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		items[i] = apiPermExclude
	}

	// Assemble next/prev page queries.
	query := make(url.Values, 1)
	if domain != "" {
		query.Set(apiutil.DomainPermissionDomainKey, domain)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/domain_permission_excludes",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

func (p *Processor) DomainPermissionExcludeRemove(
	ctx context.Context,
	acct *gtsmodel.Account,
	id string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permExclude, err := p.state.DB.GetDomainPermissionExcludeByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission exclude %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permExclude == nil {
		err := fmt.Errorf("domain permission exclude %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Delete the permission exclude.
	if err := p.state.DB.DeleteDomainPermissionExclude(ctx, permExclude.ID); err != nil {
		err := gtserror.Newf("db error deleting domain permission exclude: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPerm(ctx, permExclude, false)
}
