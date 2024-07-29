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
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// Tag contains functions for getting/creating tags in the database.
type Tag interface {
	// GetTag gets a single tag by ID
	GetTag(ctx context.Context, id string) (*gtsmodel.Tag, error)

	// GetTagByName gets a single tag using the given name.
	GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, error)

	// PutTag inserts the given tag in the database.
	PutTag(ctx context.Context, tag *gtsmodel.Tag) error

	// GetTags gets multiple tags.
	GetTags(ctx context.Context, ids []string) ([]*gtsmodel.Tag, error)

	// GetFollowedTags gets the user's followed tags.
	GetFollowedTags(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Tag, error)

	// IsAccountFollowingTag returns whether the account follows the given tag.
	IsAccountFollowingTag(ctx context.Context, accountID string, tagID string) (bool, error)

	// PutFollowedTag creates a new followed tag for a the given user.
	// If it already exists, it returns without an error.
	PutFollowedTag(ctx context.Context, accountID string, tagID string) error

	// DeleteFollowedTag deletes a followed tag for a the given user.
	// If no such followed tag exists, it returns without an error.
	DeleteFollowedTag(ctx context.Context, accountID string, tagID string) error

	// DeleteFollowedTagsByAccountID deletes all of an account's followed tags.
	DeleteFollowedTagsByAccountID(ctx context.Context, accountID string) error

	// GetAccountIDsFollowingTagIDs returns the account IDs of any followers of the given tag IDs.
	GetAccountIDsFollowingTagIDs(ctx context.Context, tagIDs []string) ([]string, error)
}
