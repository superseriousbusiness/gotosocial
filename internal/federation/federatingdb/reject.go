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
	"fmt"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (f *federatingDB) Reject(ctx context.Context, reject vocab.ActivityStreamsReject) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(reject)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("reject", i)
		l.Debug("entering Reject")
	}

	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account or federator channel wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	rejectObject := reject.GetActivityStreamsObject()
	if rejectObject == nil {
		return errors.New("Reject: no object set on vocab.ActivityStreamsReject")
	}

	for iter := rejectObject.Begin(); iter != rejectObject.End(); iter = iter.Next() {
		// check if the object is an IRI
		if iter.IsIRI() {
			// we have just the URI of whatever is being rejected, so we need to find out what it is
			rejectedObjectIRI := iter.GetIRI()
			if uris.IsFollowPath(rejectedObjectIRI) {
				// REJECT FOLLOW
				gtsFollowRequest := &gtsmodel.FollowRequest{}
				if err := f.db.GetWhere(ctx, []db.Where{{Key: "uri", Value: rejectedObjectIRI.String()}}, gtsFollowRequest); err != nil {
					return fmt.Errorf("Reject: couldn't get follow request with id %s from the database: %s", rejectedObjectIRI.String(), err)
				}

				// make sure the addressee of the original follow is the same as whatever inbox this landed in
				if gtsFollowRequest.AccountID != receivingAccount.ID {
					return errors.New("Reject: follow object account and inbox account were not the same")
				}

				if _, err := f.db.RejectFollowRequest(ctx, gtsFollowRequest.AccountID, gtsFollowRequest.TargetAccountID); err != nil {
					return err
				}

				return nil
			}
		}

		// check if iter is an AP object / type
		if iter.GetType() == nil {
			continue
		}

		if iter.GetType().GetTypeName() == ap.ActivityFollow {
			// we have the whole object so we can figure out what we're rejecting
			// REJECT FOLLOW
			asFollow, ok := iter.GetType().(vocab.ActivityStreamsFollow)
			if !ok {
				return errors.New("Reject: couldn't parse follow into vocab.ActivityStreamsFollow")
			}
			// convert the follow to something we can understand
			gtsFollow, err := f.typeConverter.ASFollowToFollow(ctx, asFollow)
			if err != nil {
				return fmt.Errorf("Reject: error converting asfollow to gtsfollow: %s", err)
			}
			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if gtsFollow.AccountID != receivingAccount.ID {
				return errors.New("Reject: follow object account and inbox account were not the same")
			}
			if _, err := f.db.RejectFollowRequest(ctx, gtsFollow.AccountID, gtsFollow.TargetAccountID); err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}
