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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Delete removes the entry with the given id.
//
// Delete is only called for federated objects. Deletes from the Social
// Protocol instead call Update to create a Tombstone.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Delete(ctx context.Context, id *url.URL) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Delete",
			"id":   id.String(),
		},
	)
	l.Debugf("received DELETE id %s", id.String())

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		// If the target account wasn't set on the context, that means this request didn't pass through the
		// API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("DELETE: target account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("DELETE: from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("DELETE: from federator channel was set on context but couldn't be parsed")
		return nil
	}

	// in a delete we only get the URI, we can't know if we have a status or a profile or something else,
	// so we have to try a few different things...
	s, err := f.db.GetStatusByURI(id.String())
	if err == nil {
		// it's a status
		l.Debugf("uri is for status with id: %s", s.ID)
		if err := f.db.DeleteByID(s.ID, &gtsmodel.Status{}); err != nil {
			return fmt.Errorf("DELETE: err deleting status: %s", err)
		}
		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsNote,
			APActivityType:   gtsmodel.ActivityStreamsDelete,
			GTSModel:         s,
			ReceivingAccount: targetAcct,
		}
	}

	a, err := f.db.GetAccountByURI(id.String())
	if err == nil {
		// it's an account
		l.Debugf("uri is for an account with id: %s", s.ID)
		if err := f.db.DeleteByID(a.ID, &gtsmodel.Account{}); err != nil {
			return fmt.Errorf("DELETE: err deleting account: %s", err)
		}
		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsProfile,
			APActivityType:   gtsmodel.ActivityStreamsDelete,
			GTSModel:         a,
			ReceivingAccount: targetAcct,
		}
	}

	return nil
}
