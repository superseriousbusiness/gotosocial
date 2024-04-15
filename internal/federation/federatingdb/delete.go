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
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"id", id},
		}...)
	l.Debug("entering Delete")

	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requestingAcct := activityContext.requestingAcct
	receivingAcct := activityContext.receivingAcct

	// in a delete we only get the URI, we can't know if we have a status or a profile or something else,
	// so we have to try a few different things...
	if s, err := f.state.DB.GetStatusByURI(ctx, id.String()); err == nil && requestingAcct.ID == s.AccountID {
		l.Debugf("deleting status: %s", s.ID)
		f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
			APObjectType:     ap.ObjectNote,
			APActivityType:   ap.ActivityDelete,
			GTSModel:         s,
			ReceivingAccount: receivingAcct,
		})
	}

	if a, err := f.state.DB.GetAccountByURI(ctx, id.String()); err == nil && requestingAcct.ID == a.ID {
		l.Debugf("deleting account: %s", a.ID)
		f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
			APObjectType:      ap.ObjectProfile,
			APActivityType:    ap.ActivityDelete,
			GTSModel:          a,
			ReceivingAccount:  receivingAcct,
			RequestingAccount: requestingAcct,
		})
	}

	return nil
}
