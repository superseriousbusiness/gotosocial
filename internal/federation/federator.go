/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federation

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/distributor"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

// Federator wraps everything needed to manage activitypub federation from gotosocial
type Federator interface {
	// Send a federated activity.
	//
	// The provided url must be the outbox of the sender. All processing of
	// the activity occurs similarly to the C2S flow:
	//   - If t is not an Activity, it is wrapped in a Create activity.
	//   - A new ID is generated for the activity.
	//   - The activity is added to the specified outbox.
	//   - The activity is prepared and delivered to recipients.
	//
	// Note that this function will only behave as expected if the
	// implementation has been constructed to support federation. This
	// method will guaranteed work for non-custom Actors. For custom actors,
	// care should be used to not call this method if only C2S is supported.
	Send(c context.Context, outbox *url.URL, t vocab.Type) (pub.Activity, error)
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
	PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (context.Context, error)
	// AuthenticatePostInbox delegates the authentication of a POST to an
	// inbox.
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
	// Blocked should determine whether to permit a set of actors given by
	// their ids are able to interact with this particular end user due to
	// being blocked or other application-specific logic.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox.
	//
	// If no error is returned, but authentication or authorization fails,
	// then blocked must be true and error nil. An http.StatusForbidden
	// will be written in the wresponse.
	//
	// Finally, if the authentication and authorization succeeds, then
	// blocked must be false and error nil. The request will continue
	// to be processed.
	Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error)
	// FederatingCallbacks returns the application logic that handles
	// ActivityStreams received from federating peers.
	//
	// Note that certain types of callbacks will be 'wrapped' with default
	// behaviors supported natively by the library. Other callbacks
	// compatible with streams.TypeResolver can be specified by 'other'.
	//
	// For example, setting the 'Create' field in the
	// FederatingWrappedCallbacks lets an application dependency inject
	// additional behaviors they want to take place, including the default
	// behavior supplied by this library. This is guaranteed to be compliant
	// with the ActivityPub Social protocol.
	//
	// To override the default behavior, instead supply the function in
	// 'other', which does not guarantee the application will be compliant
	// with the ActivityPub Social Protocol.
	//
	// Applications are not expected to handle every single ActivityStreams
	// type and extension. The unhandled ones are passed to DefaultCallback.
	FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error)
	// DefaultCallback is called for types that go-fed can deserialize but
	// are not handled by the application's callbacks returned in the
	// Callbacks method.
	//
	// Applications are not expected to handle every single ActivityStreams
	// type and extension, so the unhandled ones are passed to
	// DefaultCallback.
	DefaultCallback(c context.Context, activity pub.Activity) error
	// MaxInboxForwardingRecursionDepth determines how deep to search within
	// an activity to determine if inbox forwarding needs to occur.
	//
	// Zero or negative numbers indicate infinite recursion.
	MaxInboxForwardingRecursionDepth(c context.Context) int
	// MaxDeliveryRecursionDepth determines how deep to search within
	// collections owned by peers when they are targeted to receive a
	// delivery.
	//
	// Zero or negative numbers indicate infinite recursion.
	MaxDeliveryRecursionDepth(c context.Context) int
	// FilterForwarding allows the implementation to apply business logic
	// such as blocks, spam filtering, and so on to a list of potential
	// Collections and OrderedCollections of recipients when inbox
	// forwarding has been triggered.
	//
	// The activity is provided as a reference for more intelligent
	// logic to be used, but the implementation must not modify it.
	FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error)
	// GetInbox returns the OrderedCollection inbox of the actor for this
	// context. It is up to the implementation to provide the correct
	// collection for the kind of authorization given in the request.
	//
	// AuthenticateGetInbox will be called prior to this.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
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
	NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error)
}

type federator struct {
	actor              pub.FederatingActor
	distributor        distributor.Distributor
	federatingProtocol pub.FederatingProtocol
	commonBehavior     pub.CommonBehavior
	clock              pub.Clock
}

// NewFederator returns a new federator
func NewFederator(db db.DB, transportController transport.Controller, config *config.Config, log *logrus.Logger, distributor distributor.Distributor) Federator {

	clock := &Clock{}
	federatingProtocol := NewFederatingProtocol(db, log, config, transportController)
	commonBehavior := newCommonBehavior(db, log, config, transportController)
	actor := pub.NewFederatingActor(commonBehavior, federatingProtocol, db.Federation(), clock)

	return &federator{
		actor:              actor,
		distributor:        distributor,
		federatingProtocol: federatingProtocol,
		commonBehavior:     commonBehavior,
		clock:              clock,
	}
}

func (f *federator) Send(c context.Context, outbox *url.URL, t vocab.Type) (pub.Activity, error) {
	return f.actor.Send(c, outbox, t)
}

func (f *federator) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	return f.federatingProtocol.PostInboxRequestBodyHook(c, r, activity)
}

func (f *federator) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return f.federatingProtocol.AuthenticatePostInbox(c, w, r)
}

func (f *federator) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	return f.federatingProtocol.Blocked(c, actorIRIs)
}

func (f *federator) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	return f.federatingProtocol.FederatingCallbacks(c)
}

func (f *federator) DefaultCallback(c context.Context, activity pub.Activity) error {
	return f.federatingProtocol.DefaultCallback(c, activity)
}

func (f *federator) MaxInboxForwardingRecursionDepth(c context.Context) int {
	return f.federatingProtocol.MaxInboxForwardingRecursionDepth(c)
}

func (f *federator) MaxDeliveryRecursionDepth(c context.Context) int {
	return f.federatingProtocol.MaxDeliveryRecursionDepth(c)
}

func (f *federator) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	return f.federatingProtocol.FilterForwarding(c, potentialRecipients, a)
}

func (f *federator) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return f.federatingProtocol.GetInbox(c, r)
}

func (f *federator) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return f.commonBehavior.AuthenticateGetInbox(c, w, r)
}

func (f *federator) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return f.commonBehavior.AuthenticateGetOutbox(c, w, r)
}

func (f *federator) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return f.commonBehavior.GetOutbox(c, r)
}

func (f *federator) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	return f.commonBehavior.NewTransport(c, actorBoxIRI, gofedAgent)
}
