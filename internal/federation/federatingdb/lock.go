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
	"errors"
	"net/url"
)

// Lock takes a lock for the object at the specified id. If an error
// is returned, the lock must not have been taken.
//
// The lock must be able to succeed for an id that does not exist in
// the database. This means acquiring the lock does not guarantee the
// entry exists in the database.
//
// Locks are encouraged to be lightweight and in the Go layer, as some
// processes require tight loops acquiring and releasing locks.
//
// Used to ensure race conditions in multiple requests do not occur.
func (f *federatingDB) Lock(c context.Context, id *url.URL) (func(), error) {
	if id == nil {
		return nil, errors.New("Lock: id was nil")
	}
	unlock := f.locks.Lock(id.String())
	return unlock, nil
}
