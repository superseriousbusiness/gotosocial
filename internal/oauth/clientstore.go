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

package oauth

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/models"
)

type clientStore struct {
	db db.Basic
}

// NewClientStore returns an implementation of the oauth2 ClientStore interface, using the given db as a storage backend.
func NewClientStore(db db.Basic) oauth2.ClientStore {
	pts := &clientStore{
		db: db,
	}
	return pts
}

func (cs *clientStore) GetByID(ctx context.Context, clientID string) (oauth2.ClientInfo, error) {
	poc := &gtsmodel.Client{}
	if err := cs.db.GetByID(ctx, clientID, poc); err != nil {
		return nil, err
	}
	return models.New(poc.ID, poc.Secret, poc.Domain, poc.UserID), nil
}

func (cs *clientStore) Set(ctx context.Context, id string, cli oauth2.ClientInfo) error {
	poc := &gtsmodel.Client{
		ID:     cli.GetID(),
		Secret: cli.GetSecret(),
		Domain: cli.GetDomain(),
		UserID: cli.GetUserID(),
	}
	return cs.db.Put(ctx, poc)
}

func (cs *clientStore) Delete(ctx context.Context, id string) error {
	poc := &gtsmodel.Client{
		ID: id,
	}
	return cs.db.DeleteByID(ctx, id, poc)
}
