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
)

// Followers obtains the Followers Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "Followers",
			"actorIRI": actorIRI.String(),
		},
	)
	l.Debugf("entering FOLLOWERS function with actorIRI %s", actorIRI.String())

	acct := &gtsmodel.Account{}
	if err := f.db.GetWhere([]db.Where{{Key: "uri", Value: actorIRI.String()}}, acct); err != nil {
		return nil, fmt.Errorf("db error getting account with uri %s: %s", actorIRI.String(), err)
	}

	acctFollowers := []gtsmodel.Follow{}
	if err := f.db.GetFollowersByAccountID(acct.ID, &acctFollowers); err != nil {
		return nil, fmt.Errorf("db error getting followers for account id %s: %s", acct.ID, err)
	}

	followers = streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, follow := range acctFollowers {
		gtsFollower := &gtsmodel.Account{}
		if err := f.db.GetByID(follow.AccountID, gtsFollower); err != nil {
			return nil, fmt.Errorf("db error getting account id %s: %s", follow.AccountID, err)
		}
		uri, err := url.Parse(gtsFollower.URI)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s as url: %s", gtsFollower.URI, err)
		}
		items.AppendIRI(uri)
	}
	followers.SetActivityStreamsItems(items)
	return
}
