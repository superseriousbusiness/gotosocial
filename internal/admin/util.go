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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// stubbifyInstance renders the given instance as a stub,
// removing most information from it and marking it as
// suspended.
//
// For caller's convenience, this function returns the db
// names of all columns that are updated by it.
func stubbifyInstance(instance *gtsmodel.Instance, domainBlockID string) []string {
	instance.Title = ""
	instance.SuspendedAt = time.Now()
	instance.DomainBlockID = domainBlockID
	instance.ShortDescription = ""
	instance.Description = ""
	instance.Terms = ""
	instance.ContactEmail = ""
	instance.ContactAccountUsername = ""
	instance.ContactAccountID = ""
	instance.Version = ""

	return []string{
		"title",
		"suspended_at",
		"domain_block_id",
		"short_description",
		"description",
		"terms",
		"contact_email",
		"contact_account_username",
		"contact_account_id",
		"version",
	}
}

// rangeDomainAccounts iterates through all accounts
// originating from the given domain, and calls the
// provided range function on each account.
//
// If an error is returned while selecting accounts,
// the loop will stop and return the error.
func (a *Actions) rangeDomainAccounts(
	ctx context.Context,
	domain string,
	rangeF func(*gtsmodel.Account),
) error {
	var (
		limit = 50   // Limit selection to avoid spiking mem/cpu.
		maxID string // Start with empty string to select from top.
	)

	for {
		// Get (next) page of accounts.
		accounts, err := a.db.GetInstanceAccounts(ctx, domain, maxID, limit)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			return gtserror.Newf("db error getting instance accounts: %w", err)
		}

		if len(accounts) == 0 {
			// No accounts left, we're done.
			return nil
		}

		// Set next max ID for paging down.
		maxID = accounts[len(accounts)-1].ID

		// Call provided range function.
		for _, account := range accounts {
			rangeF(account)
		}
	}
}
