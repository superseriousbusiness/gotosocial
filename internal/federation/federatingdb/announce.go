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

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(announce)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("announce", i)
		l.Debug("entering Announce")
	}

	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	boost, isNew, err := f.typeConverter.ASAnnounceToStatus(ctx, announce)
	if err != nil {
		return fmt.Errorf("Announce: error converting announce to boost: %s", err)
	}

	if !isNew {
		// nothing to do here if this isn't a new announce
		return nil
	}

	// it's a new announce so pass it back to the processor async for dereferencing etc
	f.fedWorker.Queue(messages.FromFederator{
		APObjectType:     ap.ActivityAnnounce,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         boost,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
