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

	"codeberg.org/gruf/go-kv"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
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
			PrivateComment:     text.SanitizeToPlaintext(privateComment),
			PublicComment:      text.SanitizeToPlaintext(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new allow into the database.
		if err := p.state.DB.CreateDomainAllow(ctx, domainAllow); err != nil {
			err = gtserror.Newf("db error putting domain allow %s: %w", domain, err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}
	}

	actionID := id.NewULID()

	// Process domain allow side
	// effects asynchronously.
	if errWithCode := p.actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryDomain,
			TargetID:       domain,
			Type:           gtsmodel.AdminActionSuspend,
			AccountID:      adminAcct.ID,
			Text:           domainAllow.PrivateComment,
		},
		func(ctx context.Context) gtserror.MultiError {
			// Log start + finish.
			l := log.WithFields(kv.Fields{
				{"domain", domain},
				{"actionID", actionID},
			}...).WithContext(ctx)

			l.Info("processing domain allow side effects")
			defer func() { l.Info("finished processing domain allow side effects") }()

			return p.domainAllowSideEffects(ctx, domainAllow)
		},
	); errWithCode != nil {
		return nil, actionID, errWithCode
	}

	apiDomainAllow, errWithCode := p.apiDomainPerm(ctx, domainAllow, false)
	if errWithCode != nil {
		return nil, actionID, errWithCode
	}

	return apiDomainAllow, actionID, nil
}

func (p *Processor) domainAllowSideEffects(
	ctx context.Context,
	allow *gtsmodel.DomainAllow,
) gtserror.MultiError {
	if config.GetInstanceFederationMode() == config.InstanceFederationModeAllowlist {
		// We're running in allowlist mode,
		// so there are no side effects to
		// process here.
		return nil
	}

	// We're running in blocklist mode or
	// some similar mode which necessitates
	// domain allow side effects if a block
	// was in place when the allow was created.
	//
	// So, check if there's a block.
	block, err := p.state.DB.GetDomainBlock(ctx, allow.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs := gtserror.NewMultiError(1)
		errs.Appendf("db error getting domain block %s: %w", allow.Domain, err)
		return errs
	}

	if block == nil {
		// No block?
		// No problem!
		return nil
	}

	// There was a block, over which the new
	// allow ought to take precedence. To account
	// for this, just run side effects as though
	// the domain was being unblocked, while
	// leaving the existing block in place.
	//
	// Any accounts that were suspended by
	// the block will be unsuspended and be
	// able to interact with the instance again.
	return p.domainUnblockSideEffects(ctx, block)
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

	actionID := id.NewULID()

	// Process domain unallow side
	// effects asynchronously.
	if errWithCode := p.actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryDomain,
			TargetID:       domainAllow.Domain,
			Type:           gtsmodel.AdminActionUnsuspend,
			AccountID:      adminAcct.ID,
		},
		func(ctx context.Context) gtserror.MultiError {
			// Log start + finish.
			l := log.WithFields(kv.Fields{
				{"domain", domainAllow.Domain},
				{"actionID", actionID},
			}...).WithContext(ctx)

			l.Info("processing domain unallow side effects")
			defer func() { l.Info("finished processing domain unallow side effects") }()

			return p.domainUnallowSideEffects(ctx, domainAllow)
		},
	); errWithCode != nil {
		return nil, actionID, errWithCode
	}

	return apiDomainAllow, actionID, nil
}

func (p *Processor) domainUnallowSideEffects(
	ctx context.Context,
	allow *gtsmodel.DomainAllow,
) gtserror.MultiError {
	if config.GetInstanceFederationMode() == config.InstanceFederationModeAllowlist {
		// We're running in allowlist mode,
		// so there are no side effects to
		// process here.
		return nil
	}

	// We're running in blocklist mode or
	// some similar mode which necessitates
	// domain allow side effects if a block
	// was in place when the allow was removed.
	//
	// So, check if there's a block.
	block, err := p.state.DB.GetDomainBlock(ctx, allow.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs := gtserror.NewMultiError(1)
		errs.Appendf("db error getting domain block %s: %w", allow.Domain, err)
		return errs
	}

	if block == nil {
		// No block?
		// No problem!
		return nil
	}

	// There was a block, over which the previous
	// allow was taking precedence. Now that the
	// allow has been removed, we should put the
	// side effects of the block back in place.
	//
	// To do this, process the block side effects
	// again as though the block were freshly
	// created. This will mark all accounts from
	// the blocked domain as suspended, and clean
	// up their follows/following, media, etc.
	return p.domainBlockSideEffects(ctx, block)
}
