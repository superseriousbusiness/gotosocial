/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

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
func (f *federator) PostInboxRequestBodyHook(ctx context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	// extract any other IRIs involved in this activity
	otherInvolvedIRIs := []*url.URL{}

	// check if the Activity itself has an 'inReplyTo'
	if replyToable, ok := activity.(ap.ReplyToable); ok {
		if inReplyToURI := ap.ExtractInReplyToURI(replyToable); inReplyToURI != nil {
			otherInvolvedIRIs = append(otherInvolvedIRIs, inReplyToURI)
		}
	}

	// now check if the Object of the Activity (usually a Note or something) has an 'inReplyTo'
	if object := activity.GetActivityStreamsObject(); object != nil {
		if replyToable, ok := object.(ap.ReplyToable); ok {
			if inReplyToURI := ap.ExtractInReplyToURI(replyToable); inReplyToURI != nil {
				otherInvolvedIRIs = append(otherInvolvedIRIs, inReplyToURI)
			}
		}
	}

	// check for Tos and CCs on Activity itself
	if addressable, ok := activity.(ap.Addressable); ok {
		if ccURIs, err := ap.ExtractCCs(addressable); err == nil {
			otherInvolvedIRIs = append(otherInvolvedIRIs, ccURIs...)
		}
		if toURIs, err := ap.ExtractTos(addressable); err == nil {
			otherInvolvedIRIs = append(otherInvolvedIRIs, toURIs...)
		}
	}

	// and on the Object itself
	if object := activity.GetActivityStreamsObject(); object != nil {
		if addressable, ok := object.(ap.Addressable); ok {
			if ccURIs, err := ap.ExtractCCs(addressable); err == nil {
				otherInvolvedIRIs = append(otherInvolvedIRIs, ccURIs...)
			}
			if toURIs, err := ap.ExtractTos(addressable); err == nil {
				otherInvolvedIRIs = append(otherInvolvedIRIs, toURIs...)
			}
		}
	}

	// remove any duplicate entries in the slice we put together
	deduped := util.UniqueURIs(otherInvolvedIRIs)

	// clean any instances of the public URI since we don't care about that in this context
	cleaned := []*url.URL{}
	for _, u := range deduped {
		if !pub.IsPublic(u.String()) {
			cleaned = append(cleaned, u)
		}
	}

	withOtherInvolvedIRIs := context.WithValue(ctx, ap.ContextOtherInvolvedIRIs, cleaned)
	return withOtherInvolvedIRIs, nil
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
func (f *federator) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, bool, error) {
	l := log.WithFields(kv.Fields{
		{"useragent", r.UserAgent()},
		{"url", r.URL.String()},
	}...)
	l.Trace("received request to authenticate")

	if !uris.IsInboxPath(r.URL) {
		return nil, false, fmt.Errorf("path %s was not an inbox path", r.URL.String())
	}

	username, err := uris.ParseInboxPath(r.URL)
	if err != nil {
		return nil, false, fmt.Errorf("could not parse path %s: %s", r.URL.String(), err)
	}

	if username == "" {
		return nil, false, errors.New("username was empty")
	}

	receivingAccount, err := f.db.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		return nil, false, fmt.Errorf("could not fetch receiving account with username %s: %s", username, err)
	}

	publicKeyOwnerURI, errWithCode := f.AuthenticateFederatedRequest(ctx, receivingAccount.Username)
	if errWithCode != nil {
		switch errWithCode.Code() {
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
			// if 400, 401, or 403, obey the interface by writing the header and bailing
			w.WriteHeader(errWithCode.Code())
			return ctx, false, nil
		case http.StatusGone:
			// if the requesting account has gone (http 410) then likely
			// inbox post was a delete, we can just write 202 and leave,
			// since we didn't know about the account anyway, so we can't
			// do any further processing
			w.WriteHeader(http.StatusAccepted)
			return ctx, false, nil
		default:
			// if not, there's been a proper error
			return ctx, false, err
		}
	}

	// authentication has passed, so add an instance entry for this instance if it hasn't been done already
	i := &gtsmodel.Instance{}
	if err := f.db.GetWhere(ctx, []db.Where{{Key: "domain", Value: publicKeyOwnerURI.Host}}, i); err != nil {
		if err != db.ErrNoEntries {
			// there's been an actual error
			return ctx, false, fmt.Errorf("error getting requesting account with public key id %s: %s", publicKeyOwnerURI.String(), err)
		}

		// we don't have an entry for this instance yet so dereference it
		i, err = f.GetRemoteInstance(transport.WithFastfail(ctx), username, &url.URL{
			Scheme: publicKeyOwnerURI.Scheme,
			Host:   publicKeyOwnerURI.Host,
		})
		if err != nil {
			return nil, false, fmt.Errorf("could not dereference new remote instance %s during AuthenticatePostInbox: %s", publicKeyOwnerURI.Host, err)
		}

		// and put it in the db
		if err := f.db.Put(ctx, i); err != nil {
			return nil, false, fmt.Errorf("error inserting newly dereferenced instance %s: %s", publicKeyOwnerURI.Host, err)
		}
	}

	requestingAccount, err := f.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
		RequestingUsername: username,
		RemoteAccountID:    publicKeyOwnerURI,
	})
	if err != nil {
		return nil, false, fmt.Errorf("couldn't get requesting account %s: %s", publicKeyOwnerURI, err)
	}

	withRequesting := context.WithValue(ctx, ap.ContextRequestingAccount, requestingAccount)
	withReceiving := context.WithValue(withRequesting, ap.ContextReceivingAccount, receivingAccount)
	return withReceiving, true, nil
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
func (f *federator) Blocked(ctx context.Context, actorIRIs []*url.URL) (bool, error) {
	log.Debugf("entering BLOCKED function with IRI list: %+v", actorIRIs)

	// check domain blocks first for the given actor IRIs
	blocked, err := f.db.AreURIsBlocked(ctx, actorIRIs)
	if err != nil {
		return false, fmt.Errorf("error checking domain blocks of actorIRIs: %s", err)
	}
	if blocked {
		return blocked, nil
	}

	// check domain blocks for any other involved IRIs
	otherInvolvedIRIsI := ctx.Value(ap.ContextOtherInvolvedIRIs)
	otherInvolvedIRIs, ok := otherInvolvedIRIsI.([]*url.URL)
	if !ok {
		log.Error("other involved IRIs not set on request context")
		return false, errors.New("other involved IRIs not set on request context, so couldn't determine blocks")
	}
	blocked, err = f.db.AreURIsBlocked(ctx, otherInvolvedIRIs)
	if err != nil {
		return false, fmt.Errorf("error checking domain blocks of otherInvolvedIRIs: %s", err)
	}
	if blocked {
		return blocked, nil
	}

	// now check for user-level block from receiving against requesting account
	receivingAccountI := ctx.Value(ap.ContextReceivingAccount)
	receivingAccount, ok := receivingAccountI.(*gtsmodel.Account)
	if !ok {
		log.Error("receiving account not set on request context")
		return false, errors.New("receiving account not set on request context, so couldn't determine blocks")
	}
	requestingAccountI := ctx.Value(ap.ContextRequestingAccount)
	requestingAccount, ok := requestingAccountI.(*gtsmodel.Account)
	if !ok {
		log.Error("requesting account not set on request context")
		return false, errors.New("requesting account not set on request context, so couldn't determine blocks")
	}
	// the receiver shouldn't block the sender
	blocked, err = f.db.IsBlocked(ctx, receivingAccount.ID, requestingAccount.ID, false)
	if err != nil {
		return false, fmt.Errorf("error checking user-level blocks: %s", err)
	}
	if blocked {
		return blocked, nil
	}

	// get account IDs for other involved accounts
	var involvedAccountIDs []string
	for _, iri := range otherInvolvedIRIs {
		var involvedAccountID string
		if involvedStatus, err := f.db.GetStatusByURI(ctx, iri.String()); err == nil {
			involvedAccountID = involvedStatus.AccountID
		} else if involvedAccount, err := f.db.GetAccountByURI(ctx, iri.String()); err == nil {
			involvedAccountID = involvedAccount.ID
		}

		if involvedAccountID != "" {
			involvedAccountIDs = append(involvedAccountIDs, involvedAccountID)
		}
	}
	deduped := util.UniqueStrings(involvedAccountIDs)

	for _, involvedAccountID := range deduped {
		// the involved account shouldn't block whoever is making this request
		blocked, err = f.db.IsBlocked(ctx, involvedAccountID, requestingAccount.ID, false)
		if err != nil {
			return false, fmt.Errorf("error checking user-level otherInvolvedIRI blocks: %s", err)
		}
		if blocked {
			return blocked, nil
		}

		// whoever is receiving this request shouldn't block the involved account
		blocked, err = f.db.IsBlocked(ctx, receivingAccount.ID, involvedAccountID, false)
		if err != nil {
			return false, fmt.Errorf("error checking user-level otherInvolvedIRI blocks: %s", err)
		}
		if blocked {
			return blocked, nil
		}
	}

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
func (f *federator) FederatingCallbacks(ctx context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	wrapped = pub.FederatingWrappedCallbacks{
		// OnFollow determines what action to take for this particular callback
		// if a Follow Activity is handled.
		//
		// For our implementation, we always want to do nothing because we have internal logic for handling follows.
		OnFollow: pub.OnFollowDoNothing,
	}

	// override some default behaviors and trigger our own side effects
	other = []interface{}{
		func(ctx context.Context, undo vocab.ActivityStreamsUndo) error {
			return f.FederatingDB().Undo(ctx, undo)
		},
		func(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
			return f.FederatingDB().Accept(ctx, accept)
		},
		func(ctx context.Context, reject vocab.ActivityStreamsReject) error {
			return f.FederatingDB().Reject(ctx, reject)
		},
		func(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
			return f.FederatingDB().Announce(ctx, announce)
		},
	}

	return
}

// DefaultCallback is called for types that go-fed can deserialize but
// are not handled by the application's callbacks returned in the
// Callbacks method.
//
// Applications are not expected to handle every single ActivityStreams
// type and extension, so the unhandled ones are passed to
// DefaultCallback.
func (f *federator) DefaultCallback(ctx context.Context, activity pub.Activity) error {
	l := log.WithFields(kv.Fields{
		{"aptype", activity.GetTypeName()},
	}...)
	l.Debugf("received unhandle-able activity type so ignoring it")
	return nil
}

// MaxInboxForwardingRecursionDepth determines how deep to search within
// an activity to determine if inbox forwarding needs to occur.
//
// Zero or negative numbers indicate infinite recursion.
func (f *federator) MaxInboxForwardingRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// MaxDeliveryRecursionDepth determines how deep to search within
// collections owned by peers when they are targeted to receive a
// delivery.
//
// Zero or negative numbers indicate infinite recursion.
func (f *federator) MaxDeliveryRecursionDepth(ctx context.Context) int {
	// TODO
	return 4
}

// FilterForwarding allows the implementation to apply business logic
// such as blocks, spam filtering, and so on to a list of potential
// Collections and OrderedCollections of recipients when inbox
// forwarding has been triggered.
//
// The activity is provided as a reference for more intelligent
// logic to be used, but the implementation must not modify it.
func (f *federator) FilterForwarding(ctx context.Context, potentialRecipients []*url.URL, a pub.Activity) ([]*url.URL, error) {
	// TODO
	return []*url.URL{}, nil
}

// GetInbox returns the OrderedCollection inbox of the actor for this
// context. It is up to the implementation to provide the correct
// collection for the kind of authorization given in the request.
//
// AuthenticateGetInbox will be called prior to this.
//
// Always called, regardless whether the Federated Protocol or Social
// API is enabled.
func (f *federator) GetInbox(ctx context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// IMPLEMENTATION NOTE: For GoToSocial, we serve GETS to outboxes and inboxes through
	// the CLIENT API, not through the federation API, so we just do nothing here.
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}
