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
	"time"

	"codeberg.org/gruf/go-kv"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Returns an AdminActionF for
// domain allow side effects.
func (a *Actions) DomainAllowF(
	actionID string,
	domainAllow *gtsmodel.DomainAllow,
) ActionF {
	return func(ctx context.Context) gtserror.MultiError {
		l := log.
			WithContext(ctx).
			WithFields(kv.Fields{
				{"action", "allow"},
				{"actionID", actionID},
				{"domain", domainAllow.Domain},
			}...)

		// Log start + finish.
		l.Info("processing side effects")
		errs := a.domainAllowSideEffects(ctx, domainAllow)
		l.Info("finished processing side effects")

		return errs
	}
}

func (a *Actions) domainAllowSideEffects(
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
	block, err := a.db.GetDomainBlock(ctx, allow.Domain)
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
	return a.domainUnblockSideEffects(ctx, block)
}

// Returns an AdminActionF for
// domain unallow side effects.
func (a *Actions) DomainUnallowF(
	actionID string,
	domainAllow *gtsmodel.DomainAllow,
) ActionF {
	return func(ctx context.Context) gtserror.MultiError {
		l := log.
			WithContext(ctx).
			WithFields(kv.Fields{
				{"action", "unallow"},
				{"actionID", actionID},
				{"domain", domainAllow.Domain},
			}...)

		// Log start + finish.
		l.Info("processing side effects")
		errs := a.domainUnallowSideEffects(ctx, domainAllow)
		l.Info("finished processing side effects")

		return errs
	}
}

func (a *Actions) domainUnallowSideEffects(
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
	block, err := a.db.GetDomainBlock(ctx, allow.Domain)
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
	return a.domainBlockSideEffects(ctx, block)
}

func (a *Actions) DomainBlockF(
	actionID string,
	domainBlock *gtsmodel.DomainBlock,
) ActionF {
	return func(ctx context.Context) gtserror.MultiError {
		l := log.
			WithContext(ctx).
			WithFields(kv.Fields{
				{"action", "block"},
				{"actionID", actionID},
				{"domain", domainBlock.Domain},
			}...)

		skip, err := a.skipBlockSideEffects(ctx, domainBlock.Domain)
		if err != nil {
			return err
		}

		if skip != "" {
			l.Infof("skipping side effects: %s", skip)
			return nil
		}

		l.Info("processing side effects")
		errs := a.domainBlockSideEffects(ctx, domainBlock)
		l.Info("finished processing side effects")

		return errs
	}
}

// domainBlockSideEffects processes the side effects of a domain block:
//
//  1. Strip most info away from the instance entry for the domain.
//  2. Pass each account from the domain to the processor for deletion.
//
// It should be called asynchronously, since it can take a while when
// there are many accounts present on the given domain.
func (a *Actions) domainBlockSideEffects(
	ctx context.Context,
	block *gtsmodel.DomainBlock,
) gtserror.MultiError {
	var errs gtserror.MultiError

	// If we have an instance entry for this domain,
	// update it with the new block ID and clear all fields
	instance, err := a.db.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("db error getting instance %s: %w", block.Domain, err)
		return errs
	}

	if instance != nil {
		// We had an entry for this domain.
		columns := stubbifyInstance(instance, block.ID)
		if err := a.db.UpdateInstance(ctx, instance, columns...); err != nil {
			errs.Appendf("db error updating instance: %w", err)
			return errs
		}
	}

	// For each account that belongs to this domain,
	// process an account delete message to remove
	// that account's posts, media, etc.
	if err := a.rangeDomainAccounts(ctx, block.Domain, func(account *gtsmodel.Account) {
		if err := a.workers.Client.Process(ctx, &messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			GTSModel:       block,
			Origin:         account,
			Target:         account,
		}); err != nil {
			errs.Append(err)
		}
	}); err != nil {
		errs.Appendf("db error ranging through accounts: %w", err)
	}

	return errs
}

func (a *Actions) DomainUnblockF(
	actionID string,
	domainBlock *gtsmodel.DomainBlock,
) ActionF {
	return func(ctx context.Context) gtserror.MultiError {
		l := log.
			WithContext(ctx).
			WithFields(kv.Fields{
				{"action", "unblock"},
				{"actionID", actionID},
				{"domain", domainBlock.Domain},
			}...)

		l.Info("processing side effects")
		errs := a.domainUnblockSideEffects(ctx, domainBlock)
		l.Info("finished processing side effects")

		return errs
	}
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
func (a *Actions) domainUnblockSideEffects(
	ctx context.Context,
	block *gtsmodel.DomainBlock,
) gtserror.MultiError {
	var errs gtserror.MultiError

	// Update instance entry for this domain, if we have it.
	instance, err := a.db.GetInstance(ctx, block.Domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("db error getting instance %s: %w", block.Domain, err)
	}

	if instance != nil {
		// We had an entry, update it to signal
		// that it's no longer suspended.
		instance.SuspendedAt = time.Time{}
		instance.DomainBlockID = ""
		if err := a.db.UpdateInstance(
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
	if err := a.rangeDomainAccounts(ctx, block.Domain, func(account *gtsmodel.Account) {
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

		if err := a.db.UpdateAccount(
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

// skipBlockSideEffects checks if side effects of block creation
// should be skipped for the given domain, taking account of
// instance federation mode, and existence of any allows
// which ought to "shield" this domain from being blocked.
//
// If the caller should skip, the returned string will be non-zero
// and will be set to a reason why side effects should be skipped.
//
//   - blocklist mode + allow exists: "..." (skip)
//   - blocklist mode + no allow:     ""    (don't skip)
//   - allowlist mode + allow exists: ""    (don't skip)
//   - allowlist mode + no allow:     ""    (don't skip)
func (a *Actions) skipBlockSideEffects(
	ctx context.Context,
	domain string,
) (string, gtserror.MultiError) {
	var (
		skip string // Assume "" (don't skip).
		errs gtserror.MultiError
	)

	// Never skip block side effects in allowlist mode.
	fediMode := config.GetInstanceFederationMode()
	if fediMode == config.InstanceFederationModeAllowlist {
		return skip, errs
	}

	// We know we're in blocklist mode.
	//
	// We want to skip domain block side
	// effects if an allow is already
	// in place which overrides the block.

	// Check if an explicit allow exists for this domain.
	domainAllow, err := a.db.GetDomainAllow(ctx, domain)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		errs.Appendf("error getting domain allow: %w", err)
		return skip, errs
	}

	if domainAllow != nil {
		skip = "running in blocklist mode, and an explicit allow exists for this domain"
		return skip, errs
	}

	return skip, errs
}
