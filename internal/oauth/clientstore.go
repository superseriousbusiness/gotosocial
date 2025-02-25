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

package oauth

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/oauth2/v4/models"
)

type clientStore struct {
	state *state.State
}

// NewClientStore returns an implementation of the oauth2 ClientStore interface, using the given db as a storage backend.
func NewClientStore(state *state.State) oauth2.ClientStore {
	return &clientStore{state: state}
}

func (cs *clientStore) GetByID(ctx context.Context, clientID string) (oauth2.ClientInfo, error) {
	client, err := cs.state.DB.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return models.New(
		client.ID,
		client.Secret,
		client.Domain,
		client.UserID,
	), nil
}

func (cs *clientStore) Set(ctx context.Context, id string, cli oauth2.ClientInfo) error {
	return cs.state.DB.PutClient(ctx, &gtsmodel.Client{
		ID:     cli.GetID(),
		Secret: cli.GetSecret(),
		Domain: cli.GetDomain(),
		UserID: cli.GetUserID(),
	})
}

func (cs *clientStore) Delete(ctx context.Context, id string) error {
	return cs.state.DB.DeleteClientByID(ctx, id)
}
