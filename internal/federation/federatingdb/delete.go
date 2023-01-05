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
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Delete removes the entry with the given id.
//
// Delete is only called for federated objects. Deletes from the Social
// Protocol instead call Update to create a Tombstone.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Delete(ctx context.Context, id *url.URL) error {
	l := log.WithFields(kv.Fields{
		{"id", id},
	}...)
	l.Debug("entering Delete")

	receivingAccount, requestingAccount := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	// in a delete we only get the URI, we can't know if we have a status or a profile or something else,
	// so we have to try a few different things...
	if s, err := f.db.GetStatusByURI(ctx, id.String()); err == nil && requestingAccount.ID == s.AccountID {
		l.Debugf("uri is for STATUS with id: %s", s.ID)
		f.fedWorker.Queue(messages.FromFederator{
			APObjectType:     ap.ObjectNote,
			APActivityType:   ap.ActivityDelete,
			GTSModel:         s,
			ReceivingAccount: receivingAccount,
		})
	}

	if a, err := f.db.GetAccountByURI(ctx, id.String()); err == nil && requestingAccount.ID == a.ID {
		l.Debugf("uri is for ACCOUNT with id %s", a.ID)
		f.fedWorker.Queue(messages.FromFederator{
			APObjectType:     ap.ObjectProfile,
			APActivityType:   ap.ActivityDelete,
			GTSModel:         a,
			ReceivingAccount: receivingAccount,
		})
	}

	return nil
}
