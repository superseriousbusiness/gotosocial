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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// DomainKeysExpire iterates through all
// accounts belonging to the given domain,
// and expires the public key of each
// account found this way.
//
// The PublicKey for each account will be
// re-fetched next time a signed request
// from that account is received.
func (p *Processor) DomainKeysExpire(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	domain string,
) (string, gtserror.WithCode) {
	actionID := id.NewULID()

	// Process key expiration asynchronously.
	if errWithCode := p.actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             actionID,
			TargetCategory: gtsmodel.AdminActionCategoryDomain,
			TargetID:       domain,
			Type:           gtsmodel.AdminActionExpireKeys,
			AccountID:      adminAcct.ID,
		},
		func(ctx context.Context) gtserror.MultiError {
			return p.domainKeysExpireSideEffects(ctx, domain)
		},
	); errWithCode != nil {
		return actionID, errWithCode
	}

	return actionID, nil
}

func (p *Processor) domainKeysExpireSideEffects(ctx context.Context, domain string) gtserror.MultiError {
	var (
		expiresAt = time.Now()
		errs      gtserror.MultiError
	)

	// For each account on this domain, expire
	// the public key and update the account.
	if err := p.rangeDomainAccounts(ctx, domain, func(account *gtsmodel.Account) {
		account.PublicKeyExpiresAt = expiresAt
		if err := p.state.DB.UpdateAccount(ctx,
			account,
			"public_key_expires_at",
		); err != nil {
			errs.Appendf("db error updating account: %w", err)
		}
	}); err != nil {
		errs.Appendf("db error ranging through accounts: %w", err)
	}

	return errs
}
