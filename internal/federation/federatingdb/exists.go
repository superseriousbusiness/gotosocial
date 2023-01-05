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
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Exists returns true if the database has an entry for the specified
// id. It may not be owned by this application instance.
//
// The library makes this call only after acquiring a lock first.
//
// Implementation note: this just straight up isn't implemented, and doesn't *really* need to be either.
func (f *federatingDB) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	l := log.WithFields(kv.Fields{
		{"id", id},
	}...)
	l.Debug("entering Exists")
	return false, nil
}
