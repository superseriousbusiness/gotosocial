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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// Get returns the database entry for the specified id.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	l := log.WithFields(kv.Fields{{"id", id}}...)
	l.Debug("entering Get")

	switch {
	case uris.IsUserPath(id):
		acct, err := f.db.GetAccountByURI(ctx, id.String())
		if err != nil {
			return nil, err
		}
		return f.typeConverter.AccountToAS(ctx, acct)
	case uris.IsStatusesPath(id):
		status, err := f.db.GetStatusByURI(ctx, id.String())
		if err != nil {
			return nil, err
		}
		return f.typeConverter.StatusToAS(ctx, status)
	case uris.IsFollowersPath(id):
		return f.Followers(ctx, id)
	case uris.IsFollowingPath(id):
		return f.Following(ctx, id)
	default:
		return nil, fmt.Errorf("federatingDB: could not Get %s", id.String())
	}
}
