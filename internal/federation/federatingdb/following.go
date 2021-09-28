package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Following obtains the Following Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Following(ctx context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func":     "Following",
			"actorIRI": actorIRI.String(),
		},
	)
	l.Debugf("entering FOLLOWING function with actorIRI %s", actorIRI.String())

	var acct *gtsmodel.Account
	if util.IsUserPath(actorIRI) {
		username, err := util.ParseUserPath(actorIRI)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: error parsing user path: %s", err)
		}

		a, err := f.db.GetLocalAccountByUsername(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: db error getting account with uri %s: %s", actorIRI.String(), err)
		}

		acct = a
	} else if util.IsFollowingPath(actorIRI) {
		username, err := util.ParseFollowingPath(actorIRI)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: error parsing following path: %s", err)
		}

		a, err := f.db.GetLocalAccountByUsername(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: db error getting account with following uri %s: %s", actorIRI.String(), err)
		}

		acct = a
	} else {
		return nil, fmt.Errorf("FOLLOWING: could not parse actor IRI %s as users or following path", actorIRI.String())
	}

	acctFollowing, err := f.db.GetAccountFollows(ctx, acct.ID)
	if err != nil {
		return nil, fmt.Errorf("FOLLOWING: db error getting following for account id %s: %s", acct.ID, err)
	}

	following = streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, follow := range acctFollowing {
		if follow.Account == nil {
			followAccount, err := f.db.GetAccountByID(ctx, follow.AccountID)
			if err != nil {
				l.Errorf("FOLLOWING: db error getting account id %s: %s", follow.AccountID, err)
				continue
			}
			follow.Account = followAccount
		}

		uri, err := url.Parse(follow.Account.URI)
		if err != nil {
			return nil, fmt.Errorf("FOLLOWING: error parsing %s as url: %s", follow.Account.URI, err)
		}
		items.AppendIRI(uri)
	}
	following.SetActivityStreamsItems(items)
	return
}
