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
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

func (p *Processor) createDomainBlock(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
	obfuscate bool,
	publicComment string,
	privateComment string,
	subscriptionID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	// Check if a block already exists for this domain.
	domainBlock, err := p.state.DB.GetDomainBlock(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Something went wrong in the DB.
		err = gtserror.Newf("db error getting domain block %s: %w", domain, err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	if domainBlock == nil {
		// No block exists yet, create it.
		domainBlock = &gtsmodel.DomainBlock{
			ID:                 id.NewULID(),
			Domain:             domain,
			CreatedByAccountID: adminAcct.ID,
			PrivateComment:     text.SanitizeToPlaintext(privateComment),
			PublicComment:      text.SanitizeToPlaintext(publicComment),
			Obfuscate:          &obfuscate,
			SubscriptionID:     subscriptionID,
		}

		// Insert the new block into the database.
		if err := p.state.DB.CreateDomainBlock(ctx, domainBlock); err != nil {
			err = gtserror.Newf("db error putting domain block %s: %w", domain, err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}
	}

	actionID := id.NewULID()

	// Process domain block side
	// effects asynchronously.
	if errWithCode := p.actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryDomain,
			TargetID:       domain,
			Type:           gtsmodel.AdminActionSuspend,
			AccountID:      adminAcct.ID,
			Text:           domainBlock.PrivateComment,
		},
		func(ctx context.Context) gtserror.MultiError {
			// Log start + finish.
			l := log.WithFields(kv.Fields{
				{"domain", domain},
				{"actionID", actionID},
			}...).WithContext(ctx)

			l.Info("processing domain block side effects")
			defer func() { l.Info("finished processing domain block side effects") }()

			return p.domainBlockSideEffects(ctx, domainBlock)
		},
	); errWithCode != nil {
		return nil, actionID, errWithCode
	}

	apiDomainBlock, errWithCode := p.apiDomainPerm(ctx, domainBlock, false)
	if errWithCode != nil {
		return nil, actionID, errWithCode
	}

	return apiDomainBlock, actionID, nil
}

// domainBlockSideEffects processes the side effects of a domain block:
//
//  1. Strip most info away from the instance entry for the domain.
//  2. Pass each account from the domain to the processor for deletion.
//
// It should be called asynchronously, since it can take a while when
// there are many accounts present on the given domain.
func (p *Processor) domainBlockSideEffects(
	ctx context.Context,
	block *gtsmodel.DomainBlock,
) gtserror.MultiError {
	var errs gtserror.MultiError

	// If we have an instance entry for this domain,
	// update it with the new block ID and clear all fields
	instance, err := p.state.DB.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("db error getting instance %s: %w", block.Domain, err)
		return errs
	}

	if instance != nil {
		// We had an entry for this domain.
		columns := stubbifyInstance(instance, block.ID)
		if err := p.state.DB.UpdateInstance(ctx, instance, columns...); err != nil {
			errs.Appendf("db error updating instance: %w", err)
			return errs
		}
	}

	// For each account that belongs to this domain,
	// process an account delete message to remove
	// that account's posts, media, etc.
	if err := p.rangeDomainAccounts(ctx, block.Domain, func(account *gtsmodel.Account) {
		cMsg := messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			GTSModel:       block,
			OriginAccount:  account,
			TargetAccount:  account,
		}

		if err := p.state.Workers.ProcessFromClientAPI(ctx, cMsg); err != nil {
			errs.Append(err)
		}
	}); err != nil {
		errs.Appendf("db error ranging through accounts: %w", err)
	}

	return errs
}

func (p *Processor) deleteDomainBlock(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domainBlockID string,
) (*apimodel.DomainPermission, string, gtserror.WithCode) {
	domainBlock, err := p.state.DB.GetDomainBlockByID(ctx, domainBlockID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real error.
			err = gtserror.Newf("db error getting domain block: %w", err)
			return nil, "", gtserror.NewErrorInternalError(err)
		}

		// There are just no entries for this ID.
		err = fmt.Errorf("no domain block entry exists with ID %s", domainBlockID)
		return nil, "", gtserror.NewErrorNotFound(err, err.Error())
	}

	// Prepare the domain block to return, *before* the deletion goes through.
	apiDomainBlock, errWithCode := p.apiDomainPerm(ctx, domainBlock, false)
	if errWithCode != nil {
		return nil, "", errWithCode
	}

	// Delete the original domain block.
	if err := p.state.DB.DeleteDomainBlock(ctx, domainBlock.Domain); err != nil {
		err = gtserror.Newf("db error deleting domain block: %w", err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	actionID := id.NewULID()

	// Process domain unblock side
	// effects asynchronously.
	if errWithCode := p.actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryDomain,
			TargetID:       domainBlock.Domain,
			Type:           gtsmodel.AdminActionUnsuspend,
			AccountID:      adminAcct.ID,
		},
		func(ctx context.Context) gtserror.MultiError {
			// Log start + finish.
			l := log.WithFields(kv.Fields{
				{"domain", domainBlock.Domain},
				{"actionID", actionID},
			}...).WithContext(ctx)

			l.Info("processing domain unblock side effects")
			defer func() { l.Info("finished processing domain unblock side effects") }()

			return p.domainUnblockSideEffects(ctx, domainBlock)
		},
	); errWithCode != nil {
		return nil, actionID, errWithCode
	}

	return apiDomainBlock, actionID, nil
}

// domainUnblockSideEffects processes the side effects of undoing a
// domain block:
//
//  1. Mark instance entry as no longer suspended.
//  2. Mark each account from the domain as no longer suspended, if the
//     suspension origin corresponds to the ID of the provided domain block.
//
// It should be called asynchronously, since it can take a while when
// there are many accounts present on the given domain.
func (p *Processor) domainUnblockSideEffects(
	ctx context.Context,
	block *gtsmodel.DomainBlock,
) gtserror.MultiError {
	var errs gtserror.MultiError

	// Update instance entry for this domain, if we have it.
	instance, err := p.state.DB.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("db error getting instance %s: %w", block.Domain, err)
	}

	if instance != nil {
		// We had an entry, update it to signal
		// that it's no longer suspended.
		instance.SuspendedAt = time.Time{}
		instance.DomainBlockID = ""
		if err := p.state.DB.UpdateInstance(
			ctx,
			instance,
			"suspended_at",
			"domain_block_id",
		); err != nil {
			errs.Appendf("db error updating instance: %w", err)
			return errs
		}
	}

	// Unsuspend all accounts whose suspension origin was this domain block.
	if err := p.rangeDomainAccounts(ctx, block.Domain, func(account *gtsmodel.Account) {
		if account.SuspensionOrigin == "" || account.SuspendedAt.IsZero() {
			// Account wasn't suspended, nothing to do.
			return
		}

		if account.SuspensionOrigin != block.ID {
			// Account was suspended, but not by
			// this domain block, leave it alone.
			return
		}

		// Account was suspended by this domain
		// block, mark it as unsuspended.
		account.SuspendedAt = time.Time{}
		account.SuspensionOrigin = ""

		if err := p.state.DB.UpdateAccount(
			ctx,
			account,
			"suspended_at",
			"suspension_origin",
		); err != nil {
			errs.Appendf("db error updating account %s: %w", account.Username, err)
		}
	}); err != nil {
		errs.Appendf("db error ranging through accounts: %w", err)
	}

	return errs
}
