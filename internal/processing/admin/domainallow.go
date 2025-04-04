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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *Processor) createDomainAllow(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	// Check if an allow already exists for this domain.
	domainAllow, err := p.state.DB.GetDomainAllow(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something went wrong in the DB.
		err = gtserror.Newf("db error getting domain allow %s: %w", domain, err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	if domainAllow == nil {
		// No allow exists yet, create it.
		domainAllow = &gtsmodel.DomainAllow{
			ID:                 id.NewULID(),
			Domain:             domain,
			CreatedByAccountID: adminAcct.ID,
			PrivateComment:     text.StripHTMLFromText(privateComment),
			PublicComment:      text.StripHTMLFromText(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new allow into the database.
		if err := p.state.DB.PutDomainAllow(ctx, domainAllow); err != nil {
			err = gtserror.Newf("db error putting domain allow %s: %w", domain, err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}
	}

	// Run admin action to process
	// side effects of allow.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       domainAllow.Domain,
		Type:           gtsmodel.AdminActionUnsuspend,
		AccountID:      adminAcct.ID,
	}

	if errWithCode := p.state.AdminActions.Run(
		ctx,
		action,
		p.state.AdminActions.DomainAllowF(action.ID, domainAllow),
	); errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	apiDomainAllow, errWithCode := p.apiDomainPerm(ctx, domainAllow, false)
	if errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	return apiDomainAllow, action.ID, nil
}

func (p *Processor) updateDomainAllow(
	ctx context.Context,
	domainAllowID string,
	obfuscate *bool,
	publicComment *string,
	privateComment *string,
	subscriptionID *string,
) (*apimodel.DomainPermission, gtserror.WithCode) {
	domainAllow, err := p.state.DB.GetDomainAllowByID(ctx, domainAllowID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain allow: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain allow entry exists with ID %s", domainAllowID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	var columns []string
	if obfuscate != nil {
		domainAllow.Obfuscate = obfuscate
		columns = append(columns, "obfuscate")
	}
	if publicComment != nil {
		domainAllow.PublicComment = *publicComment
		columns = append(columns, "public_comment")
	}
	if privateComment != nil {
		domainAllow.PrivateComment = *privateComment
		columns = append(columns, "private_comment")
	}
	if subscriptionID != nil {
		domainAllow.SubscriptionID = *subscriptionID
		columns = append(columns, "subscription_id")
	}

	// Update the domain allow.
	if err := p.state.DB.UpdateDomainAllow(ctx, domainAllow, columns...); err != nil {
		err = gtserror.Newf("db error updating domain allow: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiDomainPerm(ctx, domainAllow, false)
}

func (p *Processor) deleteDomainAllow(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domainAllowID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	domainAllow, err := p.state.DB.GetDomainAllowByID(ctx, domainAllowID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain allow: %w", err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain allow entry exists with ID %s", domainAllowID)
		return nil, "", gtserror.NewErrorNotFound(err, err.Error())
	}

	// Prepare the domain allow to return, *before* the deletion goes through.
	apiDomainAllow, errWithCode := p.apiDomainPerm(ctx, domainAllow, false)
	if errWithCode != nil {
		return nil, "", errWithCode
	}

	// Delete the original domain allow.
	if err := p.state.DB.DeleteDomainAllow(ctx, domainAllow.Domain); err != nil {
		err = gtserror.Newf("db error deleting domain allow: %w", err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	// Run admin action to process
	// side effects of unallow.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       domainAllow.Domain,
		Type:           gtsmodel.AdminActionUnsuspend,
		AccountID:      adminAcct.ID,
	}

	if errWithCode := p.state.AdminActions.Run(
		ctx,
		action,
		p.state.AdminActions.DomainUnallowF(action.ID, domainAllow),
	); errWithCode != nil {
		return nil, action.ID, errWithCode
	}

	return apiDomainAllow, action.ID, nil
}
