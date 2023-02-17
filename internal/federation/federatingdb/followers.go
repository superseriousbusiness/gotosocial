package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Followers obtains the Followers Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	l := log.WithContext(ctx).
		WithFields(kv.Fields{
			{"id", actorIRI},
		}...)
	l.Debug("entering Followers")

	acct, err := f.getAccountForIRI(ctx, actorIRI)
	if err != nil {
		return nil, err
	}

	acctFollowers, err := f.db.GetAccountFollowedBy(ctx, acct.ID, false)
	if err != nil {
		return nil, fmt.Errorf("Followers: db error getting followers for account id %s: %s", acct.ID, err)
	}

	iris := []*url.URL{}
	for _, follow := range acctFollowers {
		if follow.Account == nil {
			a, err := f.db.GetAccountByID(ctx, follow.AccountID)
			if err != nil {
				errWrapped := fmt.Errorf("Followers: db error getting account id %s: %s", follow.AccountID, err)
				if err == db.ErrNoEntries {
					// no entry for this account id so it's probably been deleted and we haven't caught up yet
					l.Error(errWrapped)
					continue
				} else {
					// proper error
					return nil, errWrapped
				}
			}
			follow.Account = a
		}
		u, err := url.Parse(follow.Account.URI)
		if err != nil {
			return nil, err
		}
		iris = append(iris, u)
	}

	return f.collectIRIs(ctx, iris)
}
