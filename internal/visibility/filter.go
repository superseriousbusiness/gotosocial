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

package visibility

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Filter packages up a bunch of logic for checking whether given statuses or accounts are visible to a requester.
type Filter interface {
	// StatusVisible returns true if targetStatus is visible to requestingAccount, based on the
	// privacy settings of the status, and any blocks/mutes that might exist between the two accounts
	// or account domains, and other relevant accounts mentioned in or replied to by the status.
	StatusVisible(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error)

	// StatusesVisible calls StatusVisible for each status in the statuses slice, and returns a slice of only
	// statuses which are visible to the requestingAccount.
	StatusesVisible(ctx context.Context, statuses []*gtsmodel.Status, requestingAccount *gtsmodel.Account) ([]*gtsmodel.Status, error)

	// StatusHometimelineable returns true if targetStatus should be in the home timeline of the requesting account.
	//
	// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
	StatusHometimelineable(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error)

	// StatusPublictimelineable returns true if targetStatus should be in the public timeline of the requesting account.
	//
	// This function will call StatusVisible internally, so it's not necessary to call it beforehand.
	StatusPublictimelineable(ctx context.Context, targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error)

	// StatusBoostable returns true if targetStatus can be boosted by the requesting account.
	//
	// this function will call StatusVisible internally so it's not necessary to call it beforehand.
	StatusBoostable(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error)
}

type filter struct {
	db db.DB
}

// NewFilter returns a new Filter interface that will use the provided database.
func NewFilter(db db.DB) Filter {
	return &filter{
		db: db,
	}
}
