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

// Mention contains functions for getting/creating mentions in the database.
type Mention interface {
	// GetMention gets a single mention by ID
	GetMention(ctx context.Context, id string) (*gtsmodel.Mention, Error)

	// GetMentions gets multiple mentions.
	GetMentions(ctx context.Context, ids []string) ([]*gtsmodel.Mention, Error)
}
