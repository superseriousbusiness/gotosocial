package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Following obtains the Following Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Following(c context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "Following",
			"actorIRI": actorIRI.String(),
		},
	)
	l.Debugf("entering FOLLOWING function with actorIRI %s", actorIRI.String())

	acct := &gtsmodel.Account{}
	if util.IsUserPath(actorIRI) {
		if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: actorIRI.String()}}, acct); err != nil {
			return nil, fmt.Errorf("FOLLOWING: db error getting account with uri %s: %s", actorIRI.String(), err)
		}
	} else if util.IsFollowingPath(actorIRI) {
		if err := f.db.GetWhere([]db.Where{{Key: "following_uri", Value: actorIRI.String()}}, acct); err != nil {
			return nil, fmt.Errorf("FOLLOWING: db error getting account with following uri %s: %s", actorIRI.String(), err)
		}
	} else {
		return nil, fmt.Errorf("FOLLOWING: could not parse actor IRI %s as users or following path", actorIRI.String())
	}

	acctFollowing := []gtsmodel.Follow{}
	if err := f.db.GetAccountFollowing(acct.ID, &acctFollowing); err != nil {
		return nil, fmt.Errorf("FOLLOWING: db error getting following for account id %s: %s", acct.ID, err)
	}

	following = streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, follow := range acctFollowing {
		gtsFollowing := &gtsmodel.Account{}
		if err := f.db.GetByID(follow.AccountID, gtsFollowing); err != nil {
			return nil, fmt.Errorf("FOLLOWING: db error getting account id %s: %s", follow.AccountID, err)
		}
		uri, err := url.Parse(gtsFollowing.URI)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: error parsing %s as url: %s", gtsFollowing.URI, err)
		}
		items.AppendIRI(uri)
	}
	following.SetActivityStreamsItems(items)
	return
}
