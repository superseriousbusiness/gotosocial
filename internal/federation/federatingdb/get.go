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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// Get returns the database entry for the specified id.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	log.DebugKV(ctx, "id", id)

	switch {

	case uris.IsUserPath(id):
		acct, err := f.state.DB.GetAccountByURI(ctx, id.String())
		if err != nil {
			return nil, err
		}
		return f.converter.AccountToAS(ctx, acct)

	case uris.IsStatusesPath(id):
		status, err := f.state.DB.GetStatusByURI(ctx, id.String())
		if err != nil {
			return nil, err
		}
		return f.converter.StatusToAS(ctx, status)

	case uris.IsFollowersPath(id):
		return f.Followers(ctx, id)

	case uris.IsFollowingPath(id):
		return f.Following(ctx, id)

	case uris.IsAcceptsPath(id):
		return f.GetAccept(ctx, id)

	default:
		return nil, fmt.Errorf("federatingDB: could not Get %s", id.String())
	}
}
