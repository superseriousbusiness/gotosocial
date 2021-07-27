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
	"errors"
	"fmt"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Create adds a new entry to the database which must be able to be
// keyed by its id.
//
// Note that Activity values received from federated peers may also be
// created in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
//
// Under certain conditions and network activities, Create may be called
// multiple times for the same ActivityStreams object.
func (f *federatingDB) Create(ctx context.Context, asType vocab.Type) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Create",
			"asType": asType.GetTypeName(),
		},
	)
	m, err := streams.Serialize(asType)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	l.Debugf("received CREATE asType %s", string(b))

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		// If the target account wasn't set on the context, that means this request didn't pass through the
		// API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("CREATE: target account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("CREATE: from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("CREATE: from federator channel was set on context but couldn't be parsed")
		return nil
	}

	switch asType.GetTypeName() {
	case gtsmodel.ActivityStreamsCreate:
		// CREATE SOMETHING
		create, ok := asType.(vocab.ActivityStreamsCreate)
		if !ok {
			return errors.New("CREATE: could not convert type to create")
		}
		object := create.GetActivityStreamsObject()
		for objectIter := object.Begin(); objectIter != object.End(); objectIter = objectIter.Next() {
			switch objectIter.GetType().GetTypeName() {
			case gtsmodel.ActivityStreamsNote:
				// CREATE A NOTE
				note := objectIter.GetActivityStreamsNote()
				status, err := f.typeConverter.ASStatusToStatus(note)
				if err != nil {
					return fmt.Errorf("CREATE: error converting note to status: %s", err)
				}

				// id the status based on the time it was created
				statusID, err := id.NewULIDFromTime(status.CreatedAt)
				if err != nil {
					return err
				}
				status.ID = statusID

				if err := f.db.Put(status); err != nil {
					if _, ok := err.(db.ErrAlreadyExists); ok {
						// the status already exists in the database, which means we've already handled everything else,
						// so we can just return nil here and be done with it.
						return nil
					}
					// an actual error has happened
					return fmt.Errorf("CREATE: database error inserting status: %s", err)
				}

				fromFederatorChan <- gtsmodel.FromFederator{
					APObjectType:     gtsmodel.ActivityStreamsNote,
					APActivityType:   gtsmodel.ActivityStreamsCreate,
					GTSModel:         status,
					ReceivingAccount: targetAcct,
				}
			}
		}
	case gtsmodel.ActivityStreamsFollow:
		// FOLLOW SOMETHING
		follow, ok := asType.(vocab.ActivityStreamsFollow)
		if !ok {
			return errors.New("CREATE: could not convert type to follow")
		}

		followRequest, err := f.typeConverter.ASFollowToFollowRequest(follow)
		if err != nil {
			return fmt.Errorf("CREATE: could not convert Follow to follow request: %s", err)
		}

		newID, err := id.NewULID()
		if err != nil {
			return err
		}
		followRequest.ID = newID

		if err := f.db.Put(followRequest); err != nil {
			return fmt.Errorf("CREATE: database error inserting follow request: %s", err)
		}

		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsFollow,
			APActivityType:   gtsmodel.ActivityStreamsCreate,
			GTSModel:         followRequest,
			ReceivingAccount: targetAcct,
		}
	case gtsmodel.ActivityStreamsLike:
		// LIKE SOMETHING
		like, ok := asType.(vocab.ActivityStreamsLike)
		if !ok {
			return errors.New("CREATE: could not convert type to like")
		}

		fave, err := f.typeConverter.ASLikeToFave(like)
		if err != nil {
			return fmt.Errorf("CREATE: could not convert Like to fave: %s", err)
		}

		newID, err := id.NewULID()
		if err != nil {
			return err
		}
		fave.ID = newID

		if err := f.db.Put(fave); err != nil {
			return fmt.Errorf("CREATE: database error inserting fave: %s", err)
		}

		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsLike,
			APActivityType:   gtsmodel.ActivityStreamsCreate,
			GTSModel:         fave,
			ReceivingAccount: targetAcct,
		}
	case gtsmodel.ActivityStreamsBlock:
		// BLOCK SOMETHING
		blockable, ok := asType.(vocab.ActivityStreamsBlock)
		if !ok {
			return errors.New("CREATE: could not convert type to block")
		}

		block, err := f.typeConverter.ASBlockToBlock(blockable)
		if err != nil {
			return fmt.Errorf("CREATE: could not convert Block to gts model block")
		}

		newID, err := id.NewULID()
		if err != nil {
			return err
		}
		block.ID = newID

		if err := f.db.Put(block); err != nil {
			return fmt.Errorf("CREATE: database error inserting block: %s", err)
		}

		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsBlock,
			APActivityType:   gtsmodel.ActivityStreamsCreate,
			GTSModel:         block,
			ReceivingAccount: targetAcct,
		}
	}
	return nil
}
