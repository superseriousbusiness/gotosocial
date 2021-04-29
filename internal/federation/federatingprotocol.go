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
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// FederatingProtocol implements the go-fed federating protocol interface
type FederatingProtocol struct {
	db                  db.DB
	log                 *logrus.Logger
	config              *config.Config
	transportController transport.Controller
}

// NewFederatingProtocol returns the gotosocial implementation of the go-fed FederatingProtocol interface
func NewFederatingProtocol(db db.DB, log *logrus.Logger, config *config.Config, transportController transport.Controller) pub.FederatingProtocol {
	return &FederatingProtocol{
		db:                  db,
		log:                 log,
		config:              config,
		transportController: transportController,
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
func (f *FederatingProtocol) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":      "PostInboxRequestBodyHook",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
	})

	if activity == nil {
		err := errors.New("nil activity in PostInboxRequestBodyHook")
		l.Debug(err)
		return nil, err
	}

	if !util.IsInboxPath(r.URL) {
		err := fmt.Errorf("url %s did not correspond to inbox path", r.URL.String())
		l.Debug(err)
		return nil, err
	}

	inboxUsername, err := util.ParseInboxPath(r.URL)
	if err != nil {
		err := fmt.Errorf("could not parse username from url: %s", r.URL.String())
		l.Debug(err)
		return nil, err
	}
	l.Tracef("parsed username %s from %s", inboxUsername, r.URL.String())
	l.Tracef("signature: %s", r.Header.Get("Signature"))

	// get the gts account from the username
	inboxAccount := &gtsmodel.Account{}
	if err := f.db.GetLocalAccountByUsername(inboxUsername, inboxAccount); err != nil {
		err := fmt.Errorf("AuthenticateGetInbox: error fetching inbox account for %s from database: %s", r.URL.String(), err)
		l.Error(err)
		// return an abridged version of the error so we don't leak anything to the caller
		return nil, errors.New("database error")
	}

	ctxWithUsername := context.WithValue(ctx, util.APUsernameKey, inboxUsername)
	ctxWithAccount := context.WithValue(ctxWithUsername, util.APAccountKey, inboxAccount)
	ctxWithActivity := context.WithValue(ctxWithAccount, util.APActivityKey, activity)
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
//
// IMPLEMENTATION NOTES:
// AuthenticatePostInbox validates an incoming federation request (!!) by deriving the public key
// of the requester from the request, checking the owner of the inbox that's being requested, and doing
// some fiddling around with http signatures.
//
// A *side effect* of calling this function is that the name of the host making the request will be set
// onto the returned context, using APRequestingHostKey. If known to us already, the remote account making
// the request will also be set on the context, using APRequestingAccountKey. If not known to us already,
// the value of this key will be set to nil and the account will have to be fetched further down the line.
func (f *FederatingProtocol) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":      "AuthenticatePostInbox",
		"useragent": r.UserAgent(),
		"url":       r.URL.String(),
	})
	l.Trace("received request to authenticate")

	// account should have been set in PostInboxRequestBodyHook
	// if it's not set, we should bail because we can't do anything
	i := ctx.Value(util.APAccountKey)
	if i == nil {
		return nil, false, errors.New("could not retrieve inbox owner")
	}
	requestedAccount, ok := i.(*gtsmodel.Account)
	if !ok {
		return nil, false, errors.New("could not cast inbox owner")
	}

	v, err := httpsig.NewVerifier(r)
	if err != nil {
		return ctx, false, fmt.Errorf("could not create http sig verifier: %s", err)
	}

	requestingPublicKeyID, err := url.Parse(v.KeyId())
	if err != nil {
		return ctx, false, fmt.Errorf("could not create parse key id into a url: %s", err)
	}

	transport, err := f.transportController.NewTransport(requestedAccount.PublicKeyURI, requestedAccount.PrivateKey)
	if err != nil {
		return ctx, false, fmt.Errorf("error creating new transport: %s", err)
	}

	b, err := transport.Dereference(ctx, requestingPublicKeyID)
	if err != nil {
		return ctx, false, fmt.Errorf("error deferencing key %s: %s", requestingPublicKeyID.String(), err)
	}

	requestingPublicKey, err := getPublicKeyFromResponse(ctx, b, requestingPublicKeyID)
	if err != nil {
		return ctx, false, fmt.Errorf("error getting key %s from response %s: %s", requestingPublicKeyID.String(), string(b), err)
	}

	algo := httpsig.RSA_SHA256
	if err := v.Verify(requestingPublicKey, algo); err != nil {
		return ctx, false, fmt.Errorf("error verifying key %s: %s", requestingPublicKeyID.String(), err)
	}

	var requestingAccount *gtsmodel.Account
	a := &gtsmodel.Account{}
	if err := f.db.GetWhere("public_key_uri", requestingPublicKeyID.String(), a); err == nil {
		// we know about this account already so we can set it on the context
		requestingAccount = a
	} else {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return ctx, false, fmt.Errorf("database error finding account with public key uri %s: %s", requestingPublicKeyID.String(), err)
		}
		// do nothing here, requestingAccount will stay nil and we'll have to figure it out further down the line
	}

	// all good at this point, so just set some stuff on the context
	contextWithHost := context.WithValue(ctx, util.APRequestingHostKey, requestingPublicKeyID.Host)
	contextWithRequestingAccount := context.WithValue(contextWithHost, util.APRequestingAccountKey, requestingAccount)

	return contextWithRequestingAccount, true, nil
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
func (f *FederatingProtocol) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
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
func (f *FederatingProtocol) FederatingCallbacks(ctx context.Context) (pub.FederatingWrappedCallbacks, []interface{}, error) {
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
func (f *FederatingProtocol) DefaultCallback(ctx context.Context, activity pub.Activity) error {
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
func (f *FederatingProtocol) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 0
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *FederatingProtocol) MaxDeliveryRecursionDepth(ctx context.Context) int {
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
func (f *FederatingProtocol) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
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
func (f *FederatingProtocol) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return nil, nil
}
