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

	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/oauth2/v4"
	"code.superseriousbusiness.org/oauth2/v4/errors"
)

type clientStore struct {
	state *state.State
}

// NewClientStore returns a minimal implementation of
// oauth2.ClientStore interface, using state as storage.
//
// Only GetByID is implemented, Set and Delete are stubs.
func NewClientStore(state *state.State) oauth2.ClientStore {
	return &clientStore{state: state}
}

func (cs *clientStore) GetByID(ctx context.Context, clientID string) (oauth2.ClientInfo, error) {
	return cs.state.DB.GetApplicationByClientID(ctx, clientID)
}

func (cs *clientStore) Set(_ context.Context, _ string, _ oauth2.ClientInfo) error {
	return errors.New("func oauth2.ClientStore.Set not implemented")
}

func (cs *clientStore) Delete(_ context.Context, _ string) error {
	return errors.New("func oauth2.ClientStore.Delete not implemented")
}
