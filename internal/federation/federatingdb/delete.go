package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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

	inboxAcctI := ctx.Value(util.APAccount)
	if inboxAcctI == nil {
		l.Error("inbox account wasn't set on context")
		return nil
	}
	inboxAcct, ok := inboxAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("inbox account was set on context but couldn't be parsed")
		return nil
	}

	fromFederatorChanI := ctx.Value(util.APFromFederatorChanKey)
	if fromFederatorChanI == nil {
		l.Error("from federator channel wasn't set on context")
		return nil
	}
	fromFederatorChan, ok := fromFederatorChanI.(chan gtsmodel.FromFederator)
	if !ok {
		l.Error("from federator channel was set on context but couldn't be parsed")
		return nil
	}

	// in a delete we only get the URI, we can't know if we have a status or a profile or something else,
	// so we have to try a few different things...
	where := []db.Where{{Key: "uri", Value: id.String()}}

	s := &gtsmodel.Status{}
	if err := f.db.GetWhere(where, s); err == nil {
		// it's a status
		l.Debugf("uri is for status with id: %s", s.ID)
		if err := f.db.DeleteByID(s.ID, &gtsmodel.Status{}); err != nil {
			return fmt.Errorf("Delete: err deleting status: %s", err)
		}
		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsNote,
			APActivityType:   gtsmodel.ActivityStreamsDelete,
			GTSModel:         s,
			ReceivingAccount: inboxAcct,
		}
	}

	a := &gtsmodel.Account{}
	if err := f.db.GetWhere(where, a); err == nil {
		// it's an account
		l.Debugf("uri is for an account with id: %s", s.ID)
		if err := f.db.DeleteByID(a.ID, &gtsmodel.Account{}); err != nil {
			return fmt.Errorf("Delete: err deleting account: %s", err)
		}
		fromFederatorChan <- gtsmodel.FromFederator{
			APObjectType:     gtsmodel.ActivityStreamsProfile,
			APActivityType:   gtsmodel.ActivityStreamsDelete,
			GTSModel:         a,
			ReceivingAccount: inboxAcct,
		}
	}

	return nil
}
