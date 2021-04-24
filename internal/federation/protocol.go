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
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Federator implements the go-fed federating protocol interface
type Federator struct {
	db     db.DB
	log    *logrus.Logger
	config *config.Config
}

// NewFederator returns the gotosocial implementation of the go-fed FederatingProtocol interface
func NewFederator(db db.DB, log *logrus.Logger, config *config.Config) pub.FederatingProtocol {
	return &Federator{
		db:     db,
		log:    log,
		config: config,
	}
}

/*
	GO FED FEDERATING PROTOCOL INTERFACE
	FederatingProtocol contains behaviors an application needs to satisfy for the
	full ActivityPub S2S implementation to be supported by this library.
	It is only required if the client application wants to support the server-to-
	server, or federating, protocol.
	It is passed to the library as a dependency injection from the client
	application.
*/

// PostInboxRequestBodyHook callback after parsing the request body for a federated request
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
func (f *Federator) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	if activity == nil {
		return nil, errors.New("nil activity in PostInboxRequestBodyHook")
	}

	l := f.log.WithFields(logrus.Fields{
		"func":      "PostInboxRequestBodyHook",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
		"aptype":    activity.GetTypeName(),
	})
	l.Debugf("received inbox post request %+v", activity)

	if !util.IsInboxPath(r.URL) {
		err := fmt.Errorf("url %s did not corresponding to inbox path", r.URL.String())
		l.Debug(err)
		return nil, err
	}

	username, err := util.ParseInboxPath(r.URL)
	if err != nil {
		err := fmt.Errorf("could not parse username from url: %s", r.URL.String())
		l.Debug(err)
		return nil, err
	}
	l.Tracef("parsed username %s from %s", username, r.URL.String())

	ctxWithUsername := context.WithValue(ctx, util.APUsernameKey, username)
	ctxWithActivity := context.WithValue(ctxWithUsername, util.APActivityKey, activity)
	return ctxWithActivity, nil
}

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
func (f *Federator) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	// TODO
	return nil, false, nil
}

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
func (f *Federator) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	// TODO
	return false, nil
}

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
func (f *Federator) FederatingCallbacks(ctx context.Context) (pub.FederatingWrappedCallbacks, []interface{}, error) {
	// TODO
	return pub.FederatingWrappedCallbacks{}, nil, nil
}

// DefaultCallback is called for types that go-fed can deserialize but
// are not handled by the application's callbacks returned in the
// Callbacks method.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension, so the unhandled ones are passed to
// DefaultCallback.
func (f *Federator) DefaultCallback(ctx context.Context, activity pub.Activity) error {
	l := f.log.WithFields(logrus.Fields{
		"func":   "DefaultCallback",
		"aptype": activity.GetTypeName(),
	})
	l.Debugf("received unhandle-able activity type so ignoring it")
	return nil
}

// MaxInboxForwardingRecursionDepth determines how deep to search within
// an activity to determine if inbox forwarding needs to occur.
//
// Zero or negative numbers indicate infinite recursion.
func (f *Federator) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *Federator) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

// FilterForwarding allows the implementation to apply business logic
// such as blocks, spam filtering, and so on to a list of potential
// Collections and OrderedCollections of recipients when inbox
// forwarding has been triggered.
//
// The activity is provided as a reference for more intelligent
// logic to be used, but the implementation must not modify it.
func (f *Federator) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return nil, nil
}

// GetInbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetInbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (f *Federator) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}
