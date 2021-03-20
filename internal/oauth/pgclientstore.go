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

package oauth

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/models"
)

type pgClientStore struct {
	conn *pg.DB
}

func NewPGClientStore(conn *pg.DB) oauth2.ClientStore {
	pts := &pgClientStore{
		conn: conn,
	}
	return pts
}

func (pcs *pgClientStore) GetByID(ctx context.Context, clientID string) (oauth2.ClientInfo, error) {
	poc := &oauthClient{
		ID: clientID,
	}
	if err := pcs.conn.WithContext(ctx).Model(poc).Where("id = ?", poc.ID).Select(); err != nil {
		return nil, fmt.Errorf("error in clientstore getbyid searching for client %s: %s", clientID, err)
	}
	return models.New(poc.ID, poc.Secret, poc.Domain, poc.UserID), nil
}

func (pcs *pgClientStore) Set(ctx context.Context, id string, cli oauth2.ClientInfo) error {
	poc := &oauthClient{
		ID:     cli.GetID(),
		Secret: cli.GetSecret(),
		Domain: cli.GetDomain(),
		UserID: cli.GetUserID(),
	}
	_, err := pcs.conn.WithContext(ctx).Model(poc).OnConflict("(id) DO UPDATE").Insert()
	if err != nil {
		return fmt.Errorf("error in clientstore set: %s", err)
	}
	return nil
}

func (pcs *pgClientStore) Delete(ctx context.Context, id string) error {
	poc := &oauthClient{
		ID: id,
	}
	_, err := pcs.conn.WithContext(ctx).Model(poc).Where("id = ?", poc.ID).Delete()
	if err != nil {
		return fmt.Errorf("error in clientstore delete: %s", err)
	}
	return nil
}

type oauthClient struct {
	ID     string
	Secret string
	Domain string
	UserID string
}
