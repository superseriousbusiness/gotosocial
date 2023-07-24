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

// Tag contains functions for getting/creating tags in the database.
type Tag interface {
	// GetTag gets a single tag by ID
	GetTag(ctx context.Context, id string) (*gtsmodel.Tag, Error)

	// GetTagByName gets a single tag using the given name.
	GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, Error)

	// GetOrCreateTag returns a tag with the given name,
	// creating it in the database if it does not yet exist.
	GetOrCreateTag(ctx context.Context, name string) (*gtsmodel.Tag, Error)

	// GetTags gets multiple tags.
	GetTags(ctx context.Context, ids []string) ([]*gtsmodel.Tag, Error)
}
