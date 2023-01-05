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
)

func (f *federatingDB) Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
	l := log.Entry{}

	if log.Level() >= level.DEBUG {
		i, err := marshalItem(undo)
		if err != nil {
			return err
		}
		l = l.WithField("undo", i)
		l.Debug("entering Undo")
	}

	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	undoObject := undo.GetActivityStreamsObject()
	if undoObject == nil {
		return errors.New("UNDO: no object set on vocab.ActivityStreamsUndo")
	}

	for iter := undoObject.Begin(); iter != undoObject.End(); iter = iter.Next() {
		if iter.GetType() == nil {
			continue
		}
		switch iter.GetType().GetTypeName() {
		case ap.ActivityFollow:
			// UNDO FOLLOW
			ASFollow, ok := iter.GetType().(vocab.ActivityStreamsFollow)
			if !ok {
				return errors.New("UNDO: couldn't parse follow into vocab.ActivityStreamsFollow")
			}
			// make sure the actor owns the follow
			if !sameActor(undo.GetActivityStreamsActor(), ASFollow.GetActivityStreamsActor()) {
				return errors.New("UNDO: follow actor and activity actor not the same")
			}
			// convert the follow to something we can understand
			gtsFollow, err := f.typeConverter.ASFollowToFollow(ctx, ASFollow)
			if err != nil {
				return fmt.Errorf("UNDO: error converting asfollow to gtsfollow: %s", err)
			}
			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if gtsFollow.TargetAccountID != receivingAccount.ID {
				return errors.New("UNDO: follow object account and inbox account were not the same")
			}
			// delete any existing FOLLOW
			if err := f.db.DeleteWhere(ctx, []db.Where{{Key: "uri", Value: gtsFollow.URI}}, &gtsmodel.Follow{}); err != nil {
				return fmt.Errorf("UNDO: db error removing follow: %s", err)
			}
			// delete any existing FOLLOW REQUEST
			if err := f.db.DeleteWhere(ctx, []db.Where{{Key: "uri", Value: gtsFollow.URI}}, &gtsmodel.FollowRequest{}); err != nil {
				return fmt.Errorf("UNDO: db error removing follow request: %s", err)
			}
			l.Debug("follow undone")
			return nil
		case ap.ActivityLike:
			// UNDO LIKE
		case ap.ActivityAnnounce:
			// UNDO BOOST/REBLOG/ANNOUNCE
		case ap.ActivityBlock:
			// UNDO BLOCK
			ASBlock, ok := iter.GetType().(vocab.ActivityStreamsBlock)
			if !ok {
				return errors.New("UNDO: couldn't parse block into vocab.ActivityStreamsBlock")
			}
			// make sure the actor owns the follow
			if !sameActor(undo.GetActivityStreamsActor(), ASBlock.GetActivityStreamsActor()) {
				return errors.New("UNDO: block actor and activity actor not the same")
			}
			// convert the block to something we can understand
			gtsBlock, err := f.typeConverter.ASBlockToBlock(ctx, ASBlock)
			if err != nil {
				return fmt.Errorf("UNDO: error converting asblock to gtsblock: %s", err)
			}
			// make sure the addressee of the original block is the same as whatever inbox this landed in
			if gtsBlock.TargetAccountID != receivingAccount.ID {
				return errors.New("UNDO: block object account and inbox account were not the same")
			}
			// delete any existing BLOCK
			if err := f.db.DeleteBlockByURI(ctx, gtsBlock.URI); err != nil {
				return fmt.Errorf("UNDO: db error removing block: %s", err)
			}
			l.Debug("block undone")
			return nil
		}
	}

	return nil
}
