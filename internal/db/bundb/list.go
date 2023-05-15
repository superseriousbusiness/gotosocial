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

package bundb

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type listDB struct {
	conn  *DBConn
	state *state.State
}

func (l *listDB) GetListByID(ctx context.Context, id string) (*gtsmodel.List, error) {

}

func (l *listDB) GetLists(ctx context.Context, accountID string) ([]*gtsmodel.List, error) {

}

func (l *listDB) PutList(ctx context.Context, list *gtsmodel.List) error {

}

func (l *listDB) UpdateList(ctx context.Context, list *gtsmodel.List, columns ...string) error {

}

func (l *listDB) DeleteListByID(ctx context.Context, id string) error {

}

func (l *listDB) GetListEntryByID(ctx context.Context, id string) (*gtsmodel.ListEntry, error) {

}

func (l *listDB) GetListEntries(ctx context.Context, listID string) ([]*gtsmodel.ListEntry, error) {

}

func (l *listDB) DeleteListEntry(ctx context.Context, id string) error {

}
