/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Instance contains functions for instance-level actions (counting instance users etc.).
type Instance interface {
	// CountInstanceUsers returns the number of known accounts registered with the given domain.
	CountInstanceUsers(ctx context.Context, domain string) (int, Error)

	// CountInstanceStatuses returns the number of known statuses posted from the given domain.
	CountInstanceStatuses(ctx context.Context, domain string) (int, Error)

	// CountInstanceDomains returns the number of known instances known that the given domain federates with.
	CountInstanceDomains(ctx context.Context, domain string) (int, Error)

	// GetInstanceAccounts returns a slice of accounts from the given instance, arranged by ID.
	GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, Error)

	// GetInstancePeers returns a slice of instances that the host instance knows about.
	GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, Error)
}
