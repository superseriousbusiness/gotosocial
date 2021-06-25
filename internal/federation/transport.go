package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// NewTransport returns a new Transport on behalf of a specific actor.
//
// The actorBoxIRI will be either the inbox or outbox of an actor who is
// attempting to do the dereferencing or delivery. Any authentication
// scheme applied on the request must be based on this actor. The
// request must contain some sort of credential of the user, such as a
// HTTP Signature.
//
// The gofedAgent passed in should be used by the Transport
// implementation in the User-Agent, as well as the application-specific
// user agent string. The gofedAgent will indicate this library's use as
// well as the library's version number.
//
// Any server-wide rate-limiting that needs to occur should happen in a
// Transport implementation. This factory function allows this to be
// created, so peer servers are not DOS'd.
//
// Any retry logic should also be handled by the Transport
// implementation.
//
// Note that the library will not maintain a long-lived pointer to the
// returned Transport so that any private credentials are able to be
// garbage collected.
func (f *federator) NewTransport(ctx context.Context, actorBoxIRI *url.URL, gofedAgent string) (pub.Transport, error) {

	var username string
	var err error

	if util.IsInboxPath(actorBoxIRI) {
		username, err = util.ParseInboxPath(actorBoxIRI)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse path %s as an inbox: %s", actorBoxIRI.String(), err)
		}
	} else if util.IsOutboxPath(actorBoxIRI) {
		username, err = util.ParseOutboxPath(actorBoxIRI)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse path %s as an outbox: %s", actorBoxIRI.String(), err)
		}
	} else {
		return nil, fmt.Errorf("id %s was neither an inbox path nor an outbox path", actorBoxIRI.String())
	}

	account := &gtsmodel.Account{}
	if err := f.db.GetLocalAccountByUsername(username, account); err != nil {
		return nil, fmt.Errorf("error getting account with username %s from the db: %s", username, err)
	}

	return f.transportController.NewTransport(account.PublicKeyURI, account.PrivateKey)
}

func (f *federator) GetTransportForUser(username string) (transport.Transport, error) {
	// We need an account to use to create a transport for dereferecing something.
	// If a username has been given, we can fetch the account with that username and use it.
	// Otherwise, we can take the instance account and use those credentials to make the request.
	ourAccount := &gtsmodel.Account{}
	var u string
	if username == "" {
		u = f.config.Host
	} else {
		u = username
	}
	if err := f.db.GetLocalAccountByUsername(u, ourAccount); err != nil {
		return nil, fmt.Errorf("error getting account %s from db: %s", username, err)
	}

	transport, err := f.transportController.NewTransport(ourAccount.PublicKeyURI, ourAccount.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating transport for user %s: %s", username, err)
	}
	return transport, nil
}
