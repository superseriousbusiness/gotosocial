/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

// Instance contains functions for instance-level actions (counting instance users etc.).
type Instance interface {
	// GetUserCountForInstance returns the number of known accounts registered with the given domain.
	GetUserCountForInstance(domain string) (int, Error)

	// GetStatusCountForInstance returns the number of known statuses posted from the given domain.
	GetStatusCountForInstance(domain string) (int, Error)

	// GetDomainCountForInstance returns the number of known instances known that the given domain federates with.
	GetDomainCountForInstance(domain string) (int, Error)

	// GetAccountsForInstance returns a slice of accounts from the given instance, arranged by ID.
	GetAccountsForInstance(domain string, maxID string, limit int) ([]*gtsmodel.Account, Error)
}
