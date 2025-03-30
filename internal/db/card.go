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

// Card contains functions for getting Cards, creating Cards, and checking various other fields on Cards.
type Card interface {
	// GetCardByID fetches the Card from the database with matching id column.
	GetCardByID(ctx context.Context, id string) (*gtsmodel.Card, error)

	// PutCard stores one Card in the database.
	PutCard(ctx context.Context, Card *gtsmodel.Card) error

	// UpdateCard updates one Card in the database.
	UpdateCard(ctx context.Context, Card *gtsmodel.Card, columns ...string) error

	// DeleteCardByID deletes one Card from the database.
	DeleteCardByID(ctx context.Context, id string) error
}
