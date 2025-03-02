package pub

import (
	"context"
	"net/http"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
)

// Common contains functions required for both the Social API and Federating
// Protocol.
//
// It is passed to the library as a dependency injection from the client
// application.
type CommonBehavior interface {
	// AuthenticateGetInbox delegates the authentication of a GET to an
	// inbox.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	//
	// If an error is returned, it is passed back to the caller of
	// GetInbox. In this case, the implementation must not write a
	// response to the ResponseWriter as is expected that the client will
	// do so when handling the error. The 'authenticated' is ignored.
	//
	// If no error is returned, but authentication or authorization fails,
	// then authenticated must be false and error nil. It is expected that
	// the implementation handles writing to the ResponseWriter in this
	// case.
	//
	// Finally, if the authentication and authorization succeeds, then
	// authenticated must be true and error nil. The request will continue
	// to be processed.
	AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
	// AuthenticateGetOutbox delegates the authentication of a GET to an
	// outbox.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	//
	// If an error is returned, it is passed back to the caller of
	// GetOutbox. In this case, the implementation must not write a
	// response to the ResponseWriter as is expected that the client will
	// do so when handling the error. The 'authenticated' is ignored.
	//
	// If no error is returned, but authentication or authorization fails,
	// then authenticated must be false and error nil. It is expected that
	// the implementation handles writing to the ResponseWriter in this
	// case.
	//
	// Finally, if the authentication and authorization succeeds, then
	// authenticated must be true and error nil. The request will continue
	// to be processed.
	AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
	// GetOutbox returns the OrderedCollection inbox of the actor for this
	// context. It is up to the implementation to provide the correct
	// collection for the kind of authorization given in the request.
	//
	// AuthenticateGetOutbox will be called prior to this.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
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
	NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t Transport, err error)
}
