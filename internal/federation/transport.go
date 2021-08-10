package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-fed/activity/pub"
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

	return f.transportController.NewTransportForUsername(username)
}
