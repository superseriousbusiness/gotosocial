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
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (f *federatingDB) Undo(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Undo",
			"asType": undo.GetTypeName(),
		},
	)
	m, err := streams.Serialize(undo)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	l.Debugf("received UNDO asType %s", string(b))

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		l.Error("UNDO: target account wasn't set on context")
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("UNDO: target account was set on context but couldn't be parsed")
		return nil
	}

	undoObject := undo.GetActivityStreamsObject()
	if undoObject == nil {
		return errors.New("UNDO: no object set on vocab.ActivityStreamsUndo")
	}

	for iter := undoObject.Begin(); iter != undoObject.End(); iter = iter.Next() {
		switch iter.GetType().GetTypeName() {
		case string(gtsmodel.ActivityStreamsFollow):
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
			gtsFollow, err := f.typeConverter.ASFollowToFollow(ASFollow)
			if err != nil {
				return fmt.Errorf("UNDO: error converting asfollow to gtsfollow: %s", err)
			}
			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if gtsFollow.TargetAccountID != targetAcct.ID {
				return errors.New("UNDO: follow object account and inbox account were not the same")
			}
			// delete any existing FOLLOW
			if err := f.db.DeleteWhere([]db.Where{{Key: "uri", Value: gtsFollow.URI}}, &gtsmodel.Follow{}); err != nil {
				return fmt.Errorf("UNDO: db error removing follow: %s", err)
			}
			// delete any existing FOLLOW REQUEST
			if err := f.db.DeleteWhere([]db.Where{{Key: "uri", Value: gtsFollow.URI}}, &gtsmodel.FollowRequest{}); err != nil {
				return fmt.Errorf("UNDO: db error removing follow request: %s", err)
			}
			l.Debug("follow undone")
			return nil
		case string(gtsmodel.ActivityStreamsLike):
			// UNDO LIKE
		case string(gtsmodel.ActivityStreamsAnnounce):
			// UNDO BOOST/REBLOG/ANNOUNCE
		}
	}

	return nil
}
