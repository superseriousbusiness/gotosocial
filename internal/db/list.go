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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type List interface {
	// GetListByID gets one list with the given id.
	GetListByID(ctx context.Context, id string) (*gtsmodel.List, error)

	// GetListsForAccountID gets all lists owned by the given accountID.
	GetListsForAccountID(ctx context.Context, accountID string) ([]*gtsmodel.List, error)

	// PutList puts a new list in the database.
	PutList(ctx context.Context, list *gtsmodel.List) error

	// UpdateList updates the given list.
	// Columns is optional, if not specified all will be updated.
	UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error

	// DeleteListByID deletes one list with the given ID.
	DeleteListByID(ctx context.Context, id string) error

	// GetListEntryByID gets one list entry with the given ID.
	GetListEntryByID(ctx context.Context, id string) (*gtsmodel.ListEntry, error)

	// GetListEntries gets list entries from the given listID, using the given parameters.
	GetListEntries(ctx context.Context, listID string, maxID string, sinceID string, minID string, limit int) ([]*gtsmodel.ListEntry, error)

	// DeleteListEntry deletes one list entry with the given id.
	DeleteListEntry(ctx context.Context, id string) error

	// DeleteListEntryForFollowID deletes all list entries with the given followID.
	DeleteListEntriesForFollowID(ctx context.Context, followID string) error
}
