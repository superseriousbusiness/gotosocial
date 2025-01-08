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
)

func (a *Actions) DomainKeysExpireF(domain string) ActionF {
	return func(ctx context.Context) gtserror.MultiError {
		var (
			expiresAt = time.Now()
			errs      gtserror.MultiError
		)

		// For each account on this domain, expire
		// the public key and update the account.
		if err := a.rangeDomainAccounts(ctx, domain, func(account *gtsmodel.Account) {
			account.PublicKeyExpiresAt = expiresAt
			if err := a.db.UpdateAccount(ctx,
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
}
