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
	"encoding/json"
	"fmt"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (f *federatingDB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Announce",
		},
	)
	m, err := streams.Serialize(announce)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	l.Debugf("received ANNOUNCE %s", string(b))

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		// If the target account wasn't set on the context, that means this request didn't pass through the
		// API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("ANNOUNCE: target account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("ANNOUNCE: from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("ANNOUNCE: from federator channel was set on context but couldn't be parsed")
		return nil
	}

	boost, isNew, err := f.typeConverter.ASAnnounceToStatus(ctx, announce)
	if err != nil {
		return fmt.Errorf("ANNOUNCE: error converting announce to boost: %s", err)
	}

	if !isNew {
		// nothing to do here if this isn't a new announce
		return nil
	}

	// it's a new announce so pass it back to the processor async for dereferencing etc
	fromFederatorChan <- gtsmodel.FromFederator{
		APObjectType:     gtsmodel.ActivityStreamsAnnounce,
		APActivityType:   gtsmodel.ActivityStreamsCreate,
		GTSModel:         boost,
		ReceivingAccount: targetAcct,
	}

	return nil
}
