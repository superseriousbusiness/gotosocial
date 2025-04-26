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
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// DomainPermissionDraftGet returns one
// domain permission draft with the given id.
func (p *Processor) DomainPermissionDraftGet(
	ctx context.Context,
	id string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permDraft, err := p.state.DB.GetDomainPermissionDraftByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission draft %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permDraft == nil {
		err := fmt.Errorf("domain permission draft %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	return p.apiDomainPerm(ctx, permDraft, false)
}

// DomainPermissionDraftsGet returns a page of
// DomainPermissionDrafts with the given parameters.
func (p *Processor) DomainPermissionDraftsGet(
	ctx context.Context,
	subscriptionID string,
	domain string,
	permType gtsmodel.DomainPermissionType,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	permDrafts, err := p.state.DB.GetDomainPermissionDrafts(
		ctx,
		permType,
		subscriptionID,
		domain,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(permDrafts)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := permDrafts[count-1].ID
	hi := permDrafts[0].ID

	// Convert each perm draft to API model.
	items := make([]any, len(permDrafts))
	for i, permDraft := range permDrafts {
		apiPermDraft, err := p.apiDomainPerm(ctx, permDraft, false)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		items[i] = apiPermDraft
	}

	// Assemble next/prev page queries.
	query := make(url.Values, 3)
	if subscriptionID != "" {
		query.Set(apiutil.DomainPermissionSubscriptionIDKey, subscriptionID)
	}
	if domain != "" {
		query.Set(apiutil.DomainPermissionDomainKey, domain)
	}
	if permType != gtsmodel.DomainPermissionUnknown {
		query.Set(apiutil.DomainPermissionPermTypeKey, permType.String())
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/admin/domain_permission_drafts",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
		Query: query,
	}), nil
}

func (p *Processor) DomainPermissionDraftCreate(
	ctx context.Context,
	acct *gtsmodel.Account,
	domain string,
	permType gtsmodel.DomainPermissionType,
	obfuscate bool,
	publicComment string,
	privateComment string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permDraft := &gtsmodel.DomainPermissionDraft{
		ID:                 id.NewULID(),
		PermissionType:     permType,
		Domain:             domain,
		CreatedByAccountID: acct.ID,
		CreatedByAccount:   acct,
		PrivateComment:     privateComment,
		PublicComment:      publicComment,
		Obfuscate:          &obfuscate,
	}

	if err := p.state.DB.PutDomainPermissionDraft(ctx, permDraft); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			const text = "a domain permission draft already exists with this permission type, domain, and subscription ID"
			err := fmt.Errorf("%w: %s", err, text)
			return nil, gtserror.NewErrorConflict(err, text)
		}

		// Real error.
		err := gtserror.Newf("db error putting domain permission draft: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPerm(ctx, permDraft, false)
}

func (p *Processor) DomainPermissionDraftAccept(
	ctx context.Context,
	acct *gtsmodel.Account,
	id string,
	overwrite bool,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	permDraft, err := p.state.DB.GetDomainPermissionDraftByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission draft %s: %w", id, err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	if permDraft == nil {
		err := fmt.Errorf("domain permission draft %s not found", id)
		return nil, "", gtserror.NewErrorNotFound(err, err.Error())
	}

	var (
		// Existing permission
		// entry, if it exists.
		existing gtsmodel.DomainPermission
	)

	// Try to get existing entry.
	switch permDraft.PermissionType {
	case gtsmodel.DomainPermissionBlock:
		existing, err = p.state.DB.GetDomainBlock(
			gtscontext.SetBarebones(ctx),
			permDraft.Domain,
		)
	case gtsmodel.DomainPermissionAllow:
		existing, err = p.state.DB.GetDomainAllow(
			gtscontext.SetBarebones(ctx),
			permDraft.Domain,
		)
	}

	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission %s: %w", id, err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	// Check if we got existing entry.
	existed := !util.IsNil(existing)
	if existed && !overwrite {
		// Domain permission exists and we shouldn't
		// overwrite it, leave everything alone.
		const text = "a domain permission already exists with this permission type and domain"
		return nil, "", gtserror.NewErrorConflict(errors.New(text), text)
	}

	// Function to clean up the accepted draft, only called if
	// creating or updating permission from draft is successful.
	deleteDraft := func() {
		if err := p.state.DB.DeleteDomainPermissionDraft(ctx, permDraft.ID); err != nil {
			log.Errorf(ctx, "db error deleting domain permission draft: %v", err)
		}
	}

	if !existed {
		// Easy case, we just need to create a new domain
		// permission from the draft, and then delete it.
		var (
			new         *apimodel.DomainPermission
			actionID    string
			errWithCode gtserror.WithCode
		)

		if permDraft.PermissionType == gtsmodel.DomainPermissionBlock {
			new, actionID, errWithCode = p.createDomainBlock(
				ctx,
				acct,
				permDraft.Domain,
				*permDraft.Obfuscate,
				permDraft.PublicComment,
				permDraft.PrivateComment,
				permDraft.SubscriptionID,
			)
		}

		if permDraft.PermissionType == gtsmodel.DomainPermissionAllow {
			new, actionID, errWithCode = p.createDomainAllow(
				ctx,
				acct,
				permDraft.Domain,
				*permDraft.Obfuscate,
				permDraft.PublicComment,
				permDraft.PrivateComment,
				permDraft.SubscriptionID,
			)
		}

		// Clean up the draft
		// before returning.
		deleteDraft()

		return new, actionID, errWithCode
	}

	// Domain permission exists but we should overwrite
	// it by just updating the existing domain permission.
	// Domain can't change, so no need to re-run side effects.
	existing.SetCreatedByAccountID(permDraft.CreatedByAccountID)
	existing.SetCreatedByAccount(permDraft.CreatedByAccount)
	existing.SetPrivateComment(permDraft.PrivateComment)
	existing.SetPublicComment(permDraft.PublicComment)
	existing.SetObfuscate(permDraft.Obfuscate)
	existing.SetSubscriptionID(permDraft.SubscriptionID)

	switch dp := existing.(type) {
	case *gtsmodel.DomainBlock:
		err = p.state.DB.UpdateDomainBlock(ctx, dp)

	case *gtsmodel.DomainAllow:
		err = p.state.DB.UpdateDomainAllow(ctx, dp)
	}

	if err != nil {
		err := gtserror.Newf("db error updating existing domain permission: %w", err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	// Clean up the draft
	// before returning.
	deleteDraft()

	apiPerm, errWithCode := p.apiDomainPerm(ctx, existing, false)
	return apiPerm, "", errWithCode
}

func (p *Processor) DomainPermissionDraftRemove(
	ctx context.Context,
	acct *gtsmodel.Account,
	id string,
	excludeTarget bool,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	permDraft, err := p.state.DB.GetDomainPermissionDraftByID(ctx, id)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting domain permission draft %s: %w", id, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if permDraft == nil {
		err := fmt.Errorf("domain permission draft %s not found", id)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Delete the permission draft.
	if err := p.state.DB.DeleteDomainPermissionDraft(ctx, permDraft.ID); err != nil {
		err := gtserror.Newf("db error deleting domain permission draft: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if excludeTarget {
		// Add a domain permission exclude
		// targeting the permDraft's domain.
		_, err = p.DomainPermissionExcludeCreate(
			ctx,
			acct,
			permDraft.Domain,
			permDraft.PrivateComment,
		)
		if err != nil && !errors.Is(err, db.ErrAlreadyExists) {
			err := gtserror.Newf("db error creating domain permission exclude: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	return p.apiDomainPerm(ctx, permDraft, false)
}
