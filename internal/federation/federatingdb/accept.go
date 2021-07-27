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

func (f *federatingDB) Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
	l := f.log.WithFields(
		logrus.Fields{
			"func":   "Accept",
			"asType": accept.GetTypeName(),
		},
	)
	m, err := streams.Serialize(accept)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	l.Debugf("received ACCEPT asType %s", string(b))

	targetAcctI := ctx.Value(util.APAccount)
	if targetAcctI == nil {
		// If the target account wasn't set on the context, that means this request didn't pass through the
		// API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}
	targetAcct, ok := targetAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("ACCEPT: target account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("ACCEPT: from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("ACCEPT: from federator channel was set on context but couldn't be parsed")
		return nil
	}

	acceptObject := accept.GetActivityStreamsObject()
	if acceptObject == nil {
		return errors.New("ACCEPT: no object set on vocab.ActivityStreamsAccept")
	}

	for iter := acceptObject.Begin(); iter != acceptObject.End(); iter = iter.Next() {
		// check if the object is an IRI
		if iter.IsIRI() {
			// we have just the URI of whatever is being accepted, so we need to find out what it is
			acceptedObjectIRI := iter.GetIRI()
			if util.IsFollowPath(acceptedObjectIRI) {
				// ACCEPT FOLLOW
				gtsFollowRequest := &gtsmodel.FollowRequest{}
				if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: acceptedObjectIRI.String()}}, gtsFollowRequest); err != nil {
					return fmt.Errorf("ACCEPT: couldn't get follow request with id %s from the database: %s", acceptedObjectIRI.String(), err)
				}

				// make sure the addressee of the original follow is the same as whatever inbox this landed in
				if gtsFollowRequest.AccountID != targetAcct.ID {
					return errors.New("ACCEPT: follow object account and inbox account were not the same")
				}
				follow, err := f.db.AcceptFollowRequest(gtsFollowRequest.AccountID, gtsFollowRequest.TargetAccountID)
				if err != nil {
					return err
				}

				fromFederatorChan <- gtsmodel.FromFederator{
					APObjectType:     gtsmodel.ActivityStreamsFollow,
					APActivityType:   gtsmodel.ActivityStreamsAccept,
					GTSModel:         follow,
					ReceivingAccount: targetAcct,
				}

				return nil
			}
		}

		// check if iter is an AP object / type
		if iter.GetType() == nil {
			continue
		}
		switch iter.GetType().GetTypeName() {
		// we have the whole object so we can figure out what we're accepting
		case string(gtsmodel.ActivityStreamsFollow):
			// ACCEPT FOLLOW
			asFollow, ok := iter.GetType().(vocab.ActivityStreamsFollow)
			if !ok {
				return errors.New("ACCEPT: couldn't parse follow into vocab.ActivityStreamsFollow")
			}
			// convert the follow to something we can understand
			gtsFollow, err := f.typeConverter.ASFollowToFollow(asFollow)
			if err != nil {
				return fmt.Errorf("ACCEPT: error converting asfollow to gtsfollow: %s", err)
			}
			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if gtsFollow.AccountID != targetAcct.ID {
				return errors.New("ACCEPT: follow object account and inbox account were not the same")
			}
			follow, err := f.db.AcceptFollowRequest(gtsFollow.AccountID, gtsFollow.TargetAccountID)
			if err != nil {
				return err
			}

			fromFederatorChan <- gtsmodel.FromFederator{
				APObjectType:     gtsmodel.ActivityStreamsFollow,
				APActivityType:   gtsmodel.ActivityStreamsAccept,
				GTSModel:         follow,
				ReceivingAccount: targetAcct,
			}

			return nil
		}
	}

	return nil
}
