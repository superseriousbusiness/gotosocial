package pub

import (
	"context"
	"net/http"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
)

// DelegateActor contains the detailed interface an application must satisfy in
// order to implement the ActivityPub specification.
//
// Note that an implementation of this interface is implicitly provided in the
// calls to NewActor, NewSocialActor, and NewFederatingActor.
//
// Implementing the DelegateActor requires familiarity with the ActivityPub
// specification because it does not a strong enough abstraction for the client
// application to ignore the ActivityPub spec. It is very possible to implement
// this interface and build a foot-gun that trashes the fediverse without being
// ActivityPub compliant. Please use with due consideration.
//
// Alternatively, build an application that uses the parts of the pub library
// that do not require implementing a DelegateActor so that the ActivityPub
// implementation is completely provided out of the box.
type DelegateActor interface {
	// Hook callback after parsing the request body for a federated request
	// to the Actor's inbox.
	//
	// Can be used to set contextual information based on the Activity
	// received.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// Warning: Neither authentication nor authorization has taken place at
	// this time. Doing anything beyond setting contextual information is
	// strongly discouraged.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox. In this case, the DelegateActor implementation must not
	// write a response to the ResponseWriter as is expected that the caller
	// to PostInbox will do so when handling the error.
	PostInboxRequestBodyHook(c context.Context, r *http.Request, activity Activity) (context.Context, error)
	// Hook callback after parsing the request body for a client request
	// to the Actor's outbox.
	//
	// Can be used to set contextual information based on the
	// ActivityStreams object received.
	//
	// Only called if the Social API is enabled.
	//
	// Warning: Neither authentication nor authorization has taken place at
	// this time. Doing anything beyond setting contextual information is
	// strongly discouraged.
	//
	// If an error is returned, it is passed back to the caller of
	// PostOutbox. In this case, the DelegateActor implementation must not
	// write a response to the ResponseWriter as is expected that the caller
	// to PostOutbox will do so when handling the error.
	PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (context.Context, error)
	// AuthenticatePostInbox delegates the authentication of a POST to an
	// inbox.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox. In this case, the implementation must not write a
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
	AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
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
	// AuthorizePostInbox delegates the authorization of an activity that
	// has been sent by POST to an inbox.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox. In this case, the implementation must not write a
	// response to the ResponseWriter as is expected that the client will
	// do so when handling the error. The 'authorized' is ignored.
	//
	// If no error is returned, but authorization fails, then authorized
	// must be false and error nil. It is expected that the implementation
	// handles writing to the ResponseWriter in this case.
	//
	// Finally, if the authentication and authorization succeeds, then
	// authorized must be true and error nil. The request will continue
	// to be processed.
	AuthorizePostInbox(c context.Context, w http.ResponseWriter, activity Activity) (authorized bool, err error)
	// PostInbox delegates the side effects of adding to the inbox and
	// determining if it is a request that should be blocked.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// As a side effect, PostInbox sets the federated data in the inbox, but
	// not on its own in the database, as InboxForwarding (which is called
	// later) must decide whether it has seen this activity before in order
	// to determine whether to do the forwarding algorithm.
	//
	// If the error is ErrObjectRequired or ErrTargetRequired, then a Bad
	// Request status is sent in the response.
	PostInbox(c context.Context, inboxIRI *url.URL, activity Activity) error
	// InboxForwarding delegates inbox forwarding logic when a POST request
	// is received in the Actor's inbox.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// The delegate is responsible for determining whether to do the inbox
	// forwarding, as well as actually conducting it if it determines it
	// needs to.
	//
	// As a side effect, InboxForwarding must set the federated data in the
	// database, independently of the inbox, however it sees fit in order to
	// determine whether it has seen the activity before.
	//
	// The provided url is the inbox of the recipient of the Activity. The
	// Activity is examined for the information about who to inbox forward
	// to.
	//
	// If an error is returned, it is returned to the caller of PostInbox.
	InboxForwarding(c context.Context, inboxIRI *url.URL, activity Activity) error
	// PostOutbox delegates the logic for side effects and adding to the
	// outbox.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled. In the case of the Social API being enabled, side
	// effects of the Activity must occur.
	//
	// The delegate is responsible for adding the activity to the database's
	// general storage for independent retrieval, and not just within the
	// actor's outbox.
	//
	// If the error is ErrObjectRequired or ErrTargetRequired, then a Bad
	// Request status is sent in the response.
	//
	// Note that 'rawJSON' is an unfortunate consequence where an 'Update'
	// Activity is the only one that explicitly cares about 'null' values in
	// JSON. Since go-fed does not differentiate between 'null' values and
	// values that are simply not present, the 'rawJSON' map is ONLY needed
	// for this narrow and specific use case.
	PostOutbox(c context.Context, a Activity, outboxIRI *url.URL, rawJSON map[string]interface{}) (deliverable bool, e error)
	// AddNewIDs sets new URL ids on the activity. It also does so for all
	// 'object' properties if the Activity is a Create type.
	//
	// Only called if the Social API is enabled.
	//
	// If an error is returned, it is returned to the caller of PostOutbox.
	AddNewIDs(c context.Context, a Activity) error
	// Deliver sends a federated message. Called only if federation is
	// enabled.
	//
	// Called if the Federated Protocol is enabled.
	//
	// The provided url is the outbox of the sender. The Activity contains
	// the information about the intended recipients.
	//
	// If an error is returned, it is returned to the caller of PostOutbox.
	Deliver(c context.Context, outbox *url.URL, activity Activity) error
	// AuthenticatePostOutbox delegates the authentication and authorization
	// of a POST to an outbox.
	//
	// Only called if the Social API is enabled.
	//
	// If an error is returned, it is passed back to the caller of
	// PostOutbox. In this case, the implementation must not write a
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
	AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
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
	// WrapInCreate wraps the provided object in a Create ActivityStreams
	// activity. The provided URL is the actor's outbox endpoint.
	//
	// Only called if the Social API is enabled.
	WrapInCreate(c context.Context, value vocab.Type, outboxIRI *url.URL) (vocab.ActivityStreamsCreate, error)
	// GetOutbox returns the OrderedCollection inbox of the actor for this
	// context. It is up to the implementation to provide the correct
	// collection for the kind of authorization given in the request.
	//
	// AuthenticateGetOutbox will be called prior to this.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
	// GetInbox returns the OrderedCollection inbox of the actor for this
	// context. It is up to the implementation to provide the correct
	// collection for the kind of authorization given in the request.
	//
	// AuthenticateGetInbox will be called prior to this.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
}
