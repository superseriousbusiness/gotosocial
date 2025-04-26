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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

type List interface {
	// GetListByID gets one list with the given id.
	GetListByID(ctx context.Context, id string) (*gtsmodel.List, error)

	// GetListsByIDs fetches all lists with the provided IDs.
	GetListsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.List, error)

	// GetListsByAccountID gets all lists owned by the given accountID.
	GetListsByAccountID(ctx context.Context, accountID string) ([]*gtsmodel.List, error)

	// GetListIDsByAccountID gets the IDs of all lists owned by the given accountID.
	GetListIDsByAccountID(ctx context.Context, accountID string) ([]string, error)

	// CountListsByAccountID counts the number of lists owned by the given accountID.
	CountListsByAccountID(ctx context.Context, accountID string) (int, error)

	// GetListsContainingFollowID gets all lists that contain the given follow with ID.
	GetListsContainingFollowID(ctx context.Context, followID string) ([]*gtsmodel.List, error)

	// GetFollowIDsInList returns all the follow IDs contained within given list ID.
	GetFollowIDsInList(ctx context.Context, listID string, page *paging.Page) ([]string, error)

	// GetFollowsInList returns all the follows contained within given list ID.
	GetFollowsInList(ctx context.Context, listID string, page *paging.Page) ([]*gtsmodel.Follow, error)

	// GetAccountIDsInList return all the account IDs (follow targets) contained within given list ID.
	GetAccountIDsInList(ctx context.Context, listID string, page *paging.Page) ([]string, error)

	// GetAccountsInList return all the accounts (follow targets) contained within given list ID.
	GetAccountsInList(ctx context.Context, listID string, page *paging.Page) ([]*gtsmodel.Account, error)

	// IsAccountInListID returns whether given account with ID is in the list with ID.
	IsAccountInList(ctx context.Context, listID string, accountID string) (bool, error)

	// PopulateList ensures that the list's struct fields are populated.
	PopulateList(ctx context.Context, list *gtsmodel.List) error

	// PutList puts a new list in the database.
	PutList(ctx context.Context, list *gtsmodel.List) error

	// UpdateList updates the given list.
	// Columns is optional, if not specified all will be updated.
	UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error

	// DeleteListByID deletes one list with the given ID.
	DeleteListByID(ctx context.Context, id string) error

	// PutListEntries inserts a slice of listEntries into the database.
	// It uses a transaction to ensure no partial updates.
	PutListEntries(ctx context.Context, listEntries []*gtsmodel.ListEntry) error

	// DeleteListEntry deletes the list entry with given list ID and follow ID.
	DeleteListEntry(ctx context.Context, listID string, followID string) error

	// DeleteAllListEntryByFollow deletes all list entries with the given followIDs.
	DeleteAllListEntriesByFollows(ctx context.Context, followIDs ...string) error
}
