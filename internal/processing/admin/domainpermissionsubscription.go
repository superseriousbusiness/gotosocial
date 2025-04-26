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
	"slices"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// DomainPermissionSubscriptionGet returns one
// domain permission subscription with the given id.
func (p *Processor) DomainPermissionSubscriptionGet(
	ctx context.Context,
	id string,
) (*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	permSub, err := p.state.DB.GetDomainPermissionSubscriptionByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission subscription %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permSub == nil {
		err := fmt.Errorf("domain permission subscription %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	return p.apiDomainPermSub(ctx, permSub)
}

// DomainPermissionSubscriptionsGet returns a page of
// DomainPermissionSubscriptions with the given parameters.
func (p *Processor) DomainPermissionSubscriptionsGet(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	permSubs, err := p.state.DB.GetDomainPermissionSubscriptions(
		ctx,
		permType,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(permSubs)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := permSubs[count-1].ID
	hi := permSubs[0].ID

	// Convert each perm sub to API model.
	items := make([]any, len(permSubs))
	for i, permSub := range permSubs {
		apiPermSub, err := p.converter.DomainPermSubToAPIDomainPermSub(ctx, permSub)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		items[i] = apiPermSub
	}

	// Assemble next/prev page queries.
	query := make(url.Values, 1)
	if permType != gtsmodel.DomainPermissionUnknown {
		query.Set(apiutil.DomainPermissionPermTypeKey, permType.String())
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/domain_permission_subscriptions",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

// DomainPermissionSubscriptionsGetByPriority returns all domain permission
// subscriptions of the given permission type, in descending priority order.
func (p *Processor) DomainPermissionSubscriptionsGetByPriority(
	ctx context.Context,
	permType gtsmodel.DomainPermissionType,
) ([]*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	permSubs, err := p.state.DB.GetDomainPermissionSubscriptionsByPriority(
		ctx,
		permType,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert each perm sub to API model.
	items := make([]*apimodel.DomainPermissionSubscription, len(permSubs))
	for i, permSub := range permSubs {
		apiPermSub, err := p.converter.DomainPermSubToAPIDomainPermSub(ctx, permSub)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		items[i] = apiPermSub
	}

	return items, nil
}

func (p *Processor) DomainPermissionSubscriptionCreate(
	ctx context.Context,
	acct *gtsmodel.Account,
	priority uint8,
	title string,
	uri string,
	contentType gtsmodel.DomainPermSubContentType,
	permType gtsmodel.DomainPermissionType,
	asDraft bool,
	fetchUsername string,
	fetchPassword string,
) (*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	permSub := &gtsmodel.DomainPermissionSubscription{
		ID:                 id.NewULID(),
		Priority:           priority,
		Title:              title,
		PermissionType:     permType,
		AsDraft:            &asDraft,
		CreatedByAccountID: acct.ID,
		CreatedByAccount:   acct,
		URI:                uri,
		ContentType:        contentType,
		FetchUsername:      fetchUsername,
		FetchPassword:      fetchPassword,
	}

	err := p.state.DB.PutDomainPermissionSubscription(ctx, permSub)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// Unique constraint conflict.
			const errText = "domain permission subscription with given URI or title already exists"
			return nil, gtserror.NewErrorConflict(errors.New(errText), errText)
		}

		// Real database error.
		err := gtserror.Newf("db error putting domain permission subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPermSub(ctx, permSub)
}

func (p *Processor) DomainPermissionSubscriptionUpdate(
	ctx context.Context,
	id string,
	priority *uint8,
	title *string,
	uri *string,
	contentType *gtsmodel.DomainPermSubContentType,
	asDraft *bool,
	adoptOrphans *bool,
	fetchUsername *string,
	fetchPassword *string,
) (*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	permSub, err := p.state.DB.GetDomainPermissionSubscriptionByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission subscription %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permSub == nil {
		err := fmt.Errorf("domain permission subscription %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	columns := make([]string, 0, 7)

	if priority != nil {
		permSub.Priority = *priority
		columns = append(columns, "priority")
	}

	if title != nil {
		permSub.Title = *title
		columns = append(columns, "title")
	}

	if uri != nil {
		permSub.URI = *uri
		columns = append(columns, "uri")
	}

	if contentType != nil {
		permSub.ContentType = *contentType
		columns = append(columns, "content_type")
	}

	if asDraft != nil {
		permSub.AsDraft = asDraft
		columns = append(columns, "as_draft")
	}

	if adoptOrphans != nil {
		permSub.AdoptOrphans = adoptOrphans
		columns = append(columns, "adopt_orphans")
	}

	if fetchPassword != nil {
		permSub.FetchPassword = *fetchPassword
		columns = append(columns, "fetch_password")
	}

	if fetchUsername != nil {
		permSub.FetchUsername = *fetchUsername
		columns = append(columns, "fetch_username")
	}

	err = p.state.DB.UpdateDomainPermissionSubscription(ctx, permSub, columns...)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// Unique constraint conflict.
			const errText = "domain permission subscription with given URI or title already exists"
			return nil, gtserror.NewErrorConflict(errors.New(errText), errText)
		}

		// Real database error.
		err := gtserror.Newf("db error updating domain permission subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPermSub(ctx, permSub)
}

func (p *Processor) DomainPermissionSubscriptionRemove(
	ctx context.Context,
	id string,
	removeChildren bool,
) (*apimodel.DomainPermissionSubscription, gtserror.WithCode) {
	permSub, err := p.state.DB.GetDomainPermissionSubscriptionByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission subscription %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permSub == nil {
		err := fmt.Errorf("domain permission subscription %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Convert to API perm sub *before* doing the deletion.
	apiPermSub, errWithCode := p.apiDomainPermSub(ctx, permSub)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// TODO in next PR: if removeChildren, then remove all
	// domain permissions that are children of this domain
	// permission subscription. If not removeChildren, then
	// just unlink them by clearing their subscription ID.
	// For now just delete the domain permission subscription.
	if err := p.state.DB.DeleteDomainPermissionSubscription(ctx, id); err != nil {
		err := gtserror.Newf("db error deleting domain permission subscription: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiPermSub, nil
}

func (p *Processor) DomainPermissionSubscriptionTest(
	ctx context.Context,
	acct *gtsmodel.Account,
	id string,
) (any, gtserror.WithCode) {
	permSub, err := p.state.DB.GetDomainPermissionSubscriptionByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission subscription %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permSub == nil {
		err := fmt.Errorf("domain permission subscription %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// To process the test/dry-run correctly, we need to get
	// all domain perm subs of this type with a *higher* priority,
	// to know whether we ought to create permissions or not.
	permSubs, err := p.state.DB.GetDomainPermissionSubscriptionsByPriority(
		ctx,
		permSub.PermissionType,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Find the index of the targeted
	// subscription in the slice.
	index := slices.IndexFunc(
		permSubs,
		func(ps *gtsmodel.DomainPermissionSubscription) bool {
			return ps.ID == permSub.ID
		},
	)

	// Get a transport for calling permSub.URI.
	tsport, err := p.transport.NewTransportForUsername(ctx, acct.Username)
	if err != nil {
		err := gtserror.Newf("error getting transport: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Everything *before* the targeted
	// subscription has a higher priority.
	higherPrios := permSubs[:index]

	// Call the permSub.URI and parse a list of perms from it.
	// Any error returned here is a "real" one, not an error
	// from fetching / parsing the list.
	createdPerms, err := p.subscriptions.ProcessDomainPermissionSubscription(
		ctx,
		permSub,
		tsport,
		higherPrios,
		true, // Dry run.
	)
	if err != nil {
		err := gtserror.Newf("error doing dry-run: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// If permSub has an error set on it now,
	// we should return it to the caller.
	if permSub.Error != "" {
		return map[string]string{
			"error": permSub.Error,
		}, nil
	}

	// No error, so return the list of
	// perms that would have been created.
	apiPerms := make([]*apimodel.DomainPermission, 0, len(createdPerms))
	for _, perm := range createdPerms {
		apiPerm, errWithCode := p.apiDomainPerm(ctx, perm, false)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiPerms = append(apiPerms, apiPerm)
	}

	return apiPerms, nil
}
